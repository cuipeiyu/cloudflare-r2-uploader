package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	bucketName      = ""
	accountId       = ""
	accessKeyId     = ""
	accessKeySecret = ""
)

func main() {
	viper.SetEnvPrefix("CFR2")
	viper.AutomaticEnv()

	bucketName = viper.GetString("BUCKET")
	accountId = viper.GetString("ACCOUNT_ID")
	accessKeyId = viper.GetString("ACCESSKEY")
	accessKeySecret = viper.GetString("SECRETKEY")

	if bucketName == "" || accountId == "" || accessKeyId == "" || accessKeySecret == "" {
		log.Fatalln("unknown cloudflare config")
		return
	}

	var rootCmd = &cobra.Command{Use: "cloudflare-r2-uploader"}

	rootCmd.AddCommand(uploadCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func uploadCmd() *cobra.Command {
	upload := &cobra.Command{
		Use:              "upload",
		Short:            "upload",
		Long:             "",
		TraverseChildren: true,
		Args:             cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			force, _ := cmd.Flags().GetBool("force")

			localPath := args[0]
			remotePath := strings.TrimLeft(args[1], "/")

			r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId),
				}, nil
			})

			cfg, err := config.LoadDefaultConfig(context.TODO(),
				config.WithEndpointResolverWithOptions(r2Resolver),
				config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
			)
			if err != nil {
				log.Fatal(err)
			}

			ctx, cancelFn := context.WithTimeout(context.Background(), time.Minute)
			defer cancelFn()

			client := s3.NewFromConfig(cfg)

			log.Printf("upload \"%s\" to \"%s\"", localPath, remotePath)

			info, err := os.Stat(localPath)
			if err != nil {
				log.Fatalln(err)
			}

			if info.IsDir() {
				count := 0

				localPathAbs, _ := filepath.Abs(localPath)

				filepath.Walk(localPathAbs, func(path string, info fs.FileInfo, err error) error {
					if err != nil {
						log.Fatalln(err)
					}

					if info.IsDir() {
						return nil // keep going
					}

					file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
					if err != nil {
						log.Fatalln(err)
					}
					defer file.Close()

					key := strings.TrimPrefix(path, localPathAbs)
					key = strings.TrimPrefix(filepath.Join(remotePath, key), "/")
					log.Printf("uploading [% 4d] %s", count, key)

					_, err = client.PutObject(ctx, &s3.PutObjectInput{
						Bucket: aws.String(bucketName),
						Key:    aws.String(key),
						Body:   file,
					})
					if err != nil {
						log.Fatalln(err)
					}

					count++

					return nil
				})

				log.Printf("uploaded %d files", count)
			} else {
				skip := !force

				if !force {
					_, err = client.HeadObject(ctx, &s3.HeadObjectInput{
						Bucket: aws.String(bucketName),
						Key:    aws.String(remotePath),
					})
					if err != nil {
						if strings.Contains(err.Error(), "Not Found") {
							skip = false
						}
					}
				}

				if skip {
					log.Printf("\"%s\" is exists will be skipped", remotePath)
				} else {
					file, err := os.OpenFile(localPath, os.O_RDONLY, os.ModePerm)
					if err != nil {
						log.Fatalln(err)
					}
					defer file.Close()

					_, err = client.PutObject(ctx, &s3.PutObjectInput{
						Bucket: aws.String(bucketName),
						Key:    aws.String(remotePath),
						Body:   file,
					})
					if err != nil {
						log.Fatalln(err)
					}
				}
			}

			log.Println("complete")
		},
	}

	// force upload
	upload.Flags().Bool("force", true, "Force upload even if the file exists.")

	return upload
}
