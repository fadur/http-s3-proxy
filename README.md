# http-s3-proxy
Docker container for proxying http to S3

Inspired by https://github.com/pottava/aws-s3-proxy, with some differences:

- Handles `If-Modified-Since` HTTP header, so browser caching works
- Only does s3 proxying, leaving auth, https and such to other components, e.g. using https://caddyserver.com

## Environment Variables

Environment Variables     | Description                                       | Required | Default 
------------------------- | ------------------------------------------------- | -------- | -----------------
AWS_S3_ENDPOINT           | The S3 API endpoint                               |          | ""
AWS_S3_BUCKET             | The `S3 bucket` to be proxied with this app.      | *        | 
AWS_S3_KEY_PREFIX         | You can configure `S3 object key` prefix.         |          | -
AWS_DEFAULT_REGION        | The AWS `region` where the S3 bucket exists.      | *        | us-east-1
AWS_ACCESS_KEY_ID         | AWS `access key` for API access.                  |          | EC2 Instance Role
AWS_SECRET_ACCESS_KEY     | AWS `secret key` for API access.                  |          | EC2 Instance Role
APP_PORT                  | The port number to be assigned for listening.     |          | 80
