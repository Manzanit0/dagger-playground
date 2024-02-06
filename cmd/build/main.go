package main

import (
	"context"
	"fmt"
	"os"

	"github.com/manzanit0/dagger-playground/pkg/build"
	flag "github.com/spf13/pflag"
)

var (
	gitRepository       = flag.String("git-repository", "", "git repository to build from")
	gitRepositoryCommit = flag.String("git-repository-commit", "HEAD", "branch to checkout")
	local               = flag.Bool("local-build", false, "build from local context, not remote repository")
	dockerfilePath      = flag.String("dockerfile", "Dockerfile", "path to the dockerfile")
	awsRegion           = flag.String("aws-region", "us-east-1", "AWS region where to upload Docker image to ECR")
	awsECRURI           = flag.String("aws-ecr-uri", "", "AWS ECR URI")
	repositoryName      = flag.String("aws-ecr-repository", "delete-me", "Image repository name")
	buildArgs           = flag.StringSlice("build-arg", nil, "Dockerfile build arguments")
)

func main() {
	flag.Parse()

	if awsECRURI == nil && *awsECRURI == "" {
		fmt.Println("you need to provide an ECR URI with the -aws-ecr-uri parameter")
		os.Exit(1)
	}

	err := build.BuildAndPush(context.Background(), &build.BuildAndPushOptions{
		AwsRegion:            *awsRegion,
		AwsEcrRepositoryName: *repositoryName,
		AwsEcrURI:            *awsECRURI,
		Local:                *local,
		GitRepositoryURL:     *gitRepository,
		GitRepositoryCommit:  *gitRepositoryCommit,
		DockerfilePath:       *dockerfilePath,
		BuildArgs:            *buildArgs,
	})

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
