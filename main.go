package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"io"
	"os"
)

// StackArgs holds the project and stack names
type StackArgs struct {
	ProjectName string
	StackName   string
}

// BucketStackSpec holds the bucket name and a custom message
type BucketStackSpec struct {
	BucketName    string `bson:"bucketName"`
	CustomMessage string `bson:"customMessage"`
}

// BucketStackArgs holds the stack and spec details
type BucketStackArgs struct {
	Stack StackArgs
	Spec  BucketStackSpec
}

// BucketStackManager manages a stack
type BucketStackManager struct {
	Name  string
	Stack *auto.Stack
}

func main() {
	// Create a new router
	router := gin.New()
	// Create a new router group
	bucket := router.Group("/bucket")
	{
		// Define the routes for the group
		bucket.POST("/", BucketHandler("up"))
		bucket.DELETE("/", BucketHandler("destroy"))
		bucket.POST("/refresh", BucketHandler("refresh"))
		bucket.POST("/cancel", BucketHandler("cancel"))
	}
	// Run the router
	err := router.Run()
	if err != nil {
		panic(err)
	}
}

// NewBucketManager creates a new bucket manager
func NewBucketManager(ctx context.Context, args BucketStackArgs) (*BucketStackManager, error) {
	// This is where the pulumi code is defined for the stack, this is the code that defines the S3 bucket
	bucketDeployFunc := func(ctx *pulumi.Context) error {
		// Create a new S3 bucket
		siteBucket, err := s3.NewBucket(ctx, args.Spec.BucketName, &s3.BucketArgs{
			Bucket: pulumi.String(args.Spec.BucketName),
			Website: s3.BucketWebsiteArgs{
				IndexDocument: pulumi.String("index.html"),
			},
		})
		if err != nil {
			return err
		}
		// Define the index content
		indexContent := fmt.Sprintf(`<html><head>
  <title>S3 Automation</title><meta charset="UTF-8">
 </head>
 <body><p>Hello, thanks for being part of this!</p><p>Made with ❤️ with <a href="https://pulumi.com">Pulumi</a></p><p>Your custom message is = %s </p>
 </body></html>
`, args.Spec.CustomMessage)
		// Create a new S3 bucket object
		if _, err := s3.NewBucketObject(ctx, "index", &s3.BucketObjectArgs{
			Bucket:      siteBucket.ID(),
			Content:     pulumi.String(indexContent),
			Key:         pulumi.String("index.html"),
			ContentType: pulumi.String("text/html; charset=utf-8"),
		}); err != nil {
			return err
		}

		// Create a new S3 bucket public access block
		accessBlock, err := s3.NewBucketPublicAccessBlock(ctx, "public-access-block", &s3.BucketPublicAccessBlockArgs{
			Bucket:          siteBucket.ID(),
			BlockPublicAcls: pulumi.Bool(false),
		})
		if err != nil {
			return err
		}
		// Create a new S3 bucket policy
		if _, err := s3.NewBucketPolicy(ctx, "bucketPolicy", &s3.BucketPolicyArgs{
			Bucket: siteBucket.ID(), // refer to the bucket created earlier
			Policy: pulumi.Any(map[string]interface{}{
				"Version": "2012-10-17",
				"Statement": []map[string]interface{}{
					{
						"Effect":    "Allow",
						"Principal": "*",
						"Action": []interface{}{
							"s3:GetObject",
						},
						"Resource": []interface{}{
							pulumi.Sprintf("arn:aws:s3:::%s/*", siteBucket.ID()), // policy refers to bucket name explicitly
						},
					},
				},
			}),
		}, pulumi.DependsOn([]pulumi.Resource{accessBlock})); err != nil {
			return err
		}

		// Export the website URL
		ctx.Export("websiteUrl", siteBucket.WebsiteEndpoint)
		return nil
	}
	// This is where the stack is created and allow inline code to be executed
	s, err := auto.UpsertStackInlineSource(ctx, args.Stack.StackName, args.Stack.ProjectName, bucketDeployFunc)
	if err != nil {
		fmt.Printf("Failed to set up a workspace: %v\n", err)
		return nil, err
	}
	fmt.Printf("Created/Selected stack %q\n", args.Stack.StackName)
	// This is where the stack config is set, this is the configuration that is passed to the pulumi code
	pulumiConfig := auto.ConfigMap{
		"aws:profile": auto.ConfigValue{Value: "dev"},
		"aws:region":  auto.ConfigValue{Value: "us-east-1"},
		fmt.Sprintf("%s:customMessage", s.Name()): auto.ConfigValue{Value: args.Spec.CustomMessage},
		fmt.Sprintf("%s:bucketName", s.Name()):    auto.ConfigValue{Value: args.Spec.BucketName},
	}
	err = s.SetAllConfig(ctx, pulumiConfig)
	if err != nil {
		return nil, err
	}
	return &BucketStackManager{
		Name:  args.Stack.StackName,
		Stack: &s,
	}, nil
}

// Run executes the specified action on the stack
func (b *BucketStackManager) Run(ctx context.Context, action string) (string, error) {
	//Depending on the action, the stack will be created, updated, destroyed or canceled
	switch action {
	case "up":
		fmt.Printf("Running stack: %s\n", b.Name)
		up, err := b.Stack.Up(ctx, optup.ProgressStreams(os.Stdout))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Website URL: %s", up.Outputs["websiteUrl"].Value.(string)), nil
	case "refresh":
		fmt.Printf("Refreshing stack: %s\n", b.Name)
		_, err := b.Stack.Refresh(ctx)
		if err != nil {
			fmt.Printf("Failed to refresh stack: %s\n", b.Name)
			return "", err
		}
		return fmt.Sprintf("Stack refreshed successfully: %s", b.Name), nil
	case "destroy":
		fmt.Printf("Starting stack destroy: %s", b.Name)
		stdoutStreamer := optdestroy.ProgressStreams(os.Stdout)

		_, err := b.Stack.Destroy(ctx, stdoutStreamer)
		if err != nil {
			fmt.Printf("Failed to destroy stack: %s\n", b.Name)
			return "", err
		}
		return fmt.Sprintf("Stack successfully destroyed: %s", b.Name), nil
	case "cancel":
		fmt.Printf("Starting stack cancel: %s\n", b.Name)
		err := b.Stack.Cancel(ctx)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Stack successfully canceled: %s", b.Name), err
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

// BucketHandler handles bucket related requests
func BucketHandler(command string) func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var body BucketStackSpec
		jsonData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		err = json.Unmarshal(jsonData, &body)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		stack, err := NewBucketManager(ctx, BucketStackArgs{
			Stack: StackArgs{
				ProjectName: "pulumi_api_rest",
				StackName:   "dev",
			},
			Spec: body,
		})
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		output, err := stack.Run(ctx, command)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": output})
	}
}
