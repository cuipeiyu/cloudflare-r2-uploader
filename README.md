# cloudflare-r2-uploader

A tool to upload files to cloudflare R2 storage;


## Environment Variables

- `CFR2_BUCKET`: Cloudflare R2 Bucket
- `CFR2_ACCOUNT_ID`: Cloudflare R2 Account ID
- `CFR2_ACCESSKEY`: Cloudflare R2 Access Key
- `CFR2_SECRETKEY`: Cloudflare R2 Secret Key


## Usage

```bash
$ go install github.com/cuipeiyu/cloudflare-r2-uploader@latest

$ cloudflare-r2-uploader local_file remote_file
# or
$ cloudflare-r2-uploader local_dir remote_dir

```
