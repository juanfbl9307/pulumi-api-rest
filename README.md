# Pulumi API REST
This is a simple API REST to manage AWS S3 buckets using Pulumi and GO.

## Requirements
- [Pulumi](https://www.pulumi.com/docs/get-started/install/)
- [Go](https://golang.org/doc/install)
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html)

## How to use
1. Clone this repository
2. Run `aws configure --profile dev` to configure your AWS credentials in dev profile
2. Run `pulumi login` to authenticate with Pulumi
3. Run `go run main.go` to start the server

## Endpoints
- `POST /bucket`: Create a new bucket stack
  - Request body:
    ```json
    {
      "bucketName": "my-bucket",
      "customMessage": "some message"
    }
    ```
- `DELETE /bucket`: Delete a bucket stack


## Summary
This is a simple API REST to manage AWS S3 buckets using Pulumi and GO.
