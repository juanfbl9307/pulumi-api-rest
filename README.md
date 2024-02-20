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
- `DELETE /bucket`: Delete a bucket stack
- `POST /bucket/refresh`: Refresh the bucket stack
- `POST /bucket/cancel`: Cancel the bucket stack# pulumi-api-rest


## Summary

The provided Go code is a web server application that uses the Gin framework for routing and the Pulumi SDK for managing AWS S3 buckets.

The application defines several types at the beginning. `StackArgs` holds the project and stack names. `BucketStackSpec` holds the bucket name and a custom message. `BucketStackArgs` holds the stack and spec details. `BucketStackManager` manages a stack and contains the stack name and a reference to the stack itself.

The `main` function initializes a new Gin router and defines a group of routes under the "/bucket" path. These routes include POST "/", DELETE "/", POST "/refresh", and POST "/cancel". Each route is associated with a handler function that performs a specific action on an AWS S3 bucket.

The `NewBucketManager` function creates a new bucket manager. It defines a Pulumi stack that creates an S3 bucket, sets up a website for the bucket, creates a bucket object for the index page, sets up a public access block, and creates a bucket policy. The function then creates or selects the stack, sets the stack configuration, and returns a new `BucketStackManager`.

The `Run` method of the `BucketStackManager` type executes a specified action on the stack. Depending on the action, the stack will be created, updated, destroyed, or canceled.

The `BucketHandler` function handles bucket-related requests. It reads the request body, unmarshals the JSON data into a `BucketStackSpec` type, creates a new bucket manager, runs the specified command on the stack, and returns the output or an error.

Here's a small code snippet to illustrate the `BucketHandler` function:

```go
func BucketHandler(command string) func(c *gin.Context) {
 return func(c *gin.Context) {
  // ...
  stack, err := NewBucketManager(ctx, BucketStackArgs{
   Stack: StackArgs{
    ProjectName: "pulumi_api_rest",
    StackName:   "dev",
   },
   Spec: body,
  })
  // ...
  output, err := stack.Run(ctx, command)
  // ...
  c.JSON(200, gin.H{"message": output})
 }
}
```

In this snippet, the `BucketHandler` function creates a new bucket manager and runs a command on the stack. The output of the command is then returned in the response.