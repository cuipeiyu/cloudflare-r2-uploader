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
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalln("usage: cloudflare-r2-uploader <local-path> <remote-path>")
		return
	}

	localPath := os.Args[1]
	remotePath := strings.TrimLeft(os.Args[2], "/")

	var bucketName = getEnv("CFR2_BUCKET", "")
	var accountId = getEnv("CFR2_ACCOUNT_ID", "")
	var accessKeyId = getEnv("CFR2_ACCESSKEY", "")
	var accessKeySecret = getEnv("CFR2_SECRETKEY", "")

	if bucketName == "" || accountId == "" || accessKeyId == "" || accessKeySecret == "" {
		log.Fatalln("unknown cloudflare config")
		return
	}

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

	log.Println("complete")
}
