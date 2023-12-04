package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"dagger.io/dagger"
)

var (
	dockerfilePath = flag.String("dockerfile", "Dockerfile", "path to the dockerfile")
	awsRegion      = flag.String("aws-region", "us-east-1", "AWS region where to upload Docker image to ECR")
	awsECRURI      = flag.String("aws-ecr-uri", "", "AWS ECR URI")
	repositoryName = flag.String("repository", "test-dagger", "Image repository name")
)

func main() {
	flag.Parse()

	if awsECRURI == nil && *awsECRURI == "" {
		panic("you need to provide an ECR URI with the -aws-ecr-uri parameter")
	}

	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	awsClient, err := NewAWSClient(ctx, *awsRegion)
	if err != nil {
		panic(err)
	}

	registry := InitRegistry(ctx, awsClient, *awsECRURI)

	exists, err := awsClient.EnsureRepositoryExists(ctx, *repositoryName)
	if err != nil {
		panic(err)
	}

	if exists {
		fmt.Println("Skipped ECR repository creation, already exists")
	} else {
		fmt.Println("Created ECR repository")
	}

	contextDir := client.Host().Directory(".")

	dockerfile := client.Host().File(*dockerfilePath)

	workspace := contextDir.WithFile("Dockerfile", dockerfile)

	ref, err := client.
		Container().
		WithRegistryAuth(
			registry.uri,
			registry.username,
			client.SetSecret("registryPassword", registry.password),
		).
		Build(workspace, dagger.ContainerBuildOpts{
			Dockerfile: "Dockerfile",
			BuildArgs: []dagger.BuildArg{
				{Name: "SERVICE_NAME", Value: "foo"},
				{Name: "SERVICE_VERSION", Value: "v1-daggerbuild"},
			},
		}).
		Publish(ctx, registry.uri+"/"+*repositoryName)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Published image to :%s\n", ref)
}
