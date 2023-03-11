# cloudfare-r2-uploader

A tool to upload files to Cloudfare R2 storage;


## Environment Variables

- `CFR2_BUCKET`: Cloudfare R2 Bucket
- `CFR2_ACCOUNT_ID`: Cloudfare R2 Account ID
- `CFR2_ACCESSKEY`: Cloudfare R2 Access Key
- `CFR2_SECRETKEY`: Cloudfare R2 Secret Key


## Usage

```bash
$ go install github.com/cuipeiyu/cloudfare-r2-uploader@latest

$ cloudfare-r2-uploader local_file remote_file
# or
$ cloudflare-r2-uploader local_dir remote_dir

```
