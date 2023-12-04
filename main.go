package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/manzanit0/dagger-playground/pkg/aws"
	flag "github.com/spf13/pflag"

	"dagger.io/dagger"
)

var (
	gitRepository       = flag.String("git-repository", "", "git repository to build from")
	gitRepositoryBranch = flag.String("git-repository-branch", "main", "branch to checkout")
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
		panic("you need to provide an ECR URI with the -aws-ecr-uri parameter")
	}

	args, err := parseBuildArgs(buildArgs)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	awsClient, err := aws.NewClient(ctx, *awsRegion)
	if err != nil {
		panic(err)
	}

	username, password, err := awsClient.GetECRUsernamePassword(ctx)
	if err != nil {
		panic(err)
	}

	exists, err := awsClient.EnsureRepositoryExists(ctx, *repositoryName)
	if err != nil {
		panic(err)
	}

	if exists {
		fmt.Println("Skipped ECR repository creation, already exists")
	} else {
		fmt.Println("Created ECR repository")
	}

	var workspace *dagger.Directory
	if *local {
		workspace = prepareLocalWorkspace(client, *dockerfilePath)
	} else {
		workspace = prepareRemoteWorkspace(client, *gitRepository, *gitRepositoryBranch, *dockerfilePath)
	}

	ref, err := client.
		Container().
		WithRegistryAuth(
			*awsECRURI,
			username,
			client.SetSecret("registryPassword", password),
		).
		Build(workspace, dagger.ContainerBuildOpts{
			Dockerfile: "Dockerfile",
			BuildArgs:  args,
		}).
		Publish(ctx, *awsECRURI+"/"+*repositoryName)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Published image to :%s\n", ref)
}

func prepareLocalWorkspace(client *dagger.Client, dockerfilePath string) *dagger.Directory {
	contextDir := client.Host().Directory(".")
	dockerfile := client.Host().File(dockerfilePath)
	return contextDir.WithFile("Dockerfile", dockerfile)
}

func prepareRemoteWorkspace(client *dagger.Client, repository, branch, dockerfilePath string) *dagger.Directory {
	// Retrieve path of authentication agent socket from host
	sshAgentPath := os.Getenv("SSH_AUTH_SOCK")

	contextDir := client.
		Git(repository, dagger.GitOpts{
			SSHAuthSocket: client.Host().UnixSocket(sshAgentPath),
		}).
		Branch(branch).
		Tree()

	dockerfile := contextDir.File(dockerfilePath)
	return contextDir.WithFile("Dockerfile", dockerfile)
}

func parseBuildArgs(argss *[]string) ([]dagger.BuildArg, error) {
	var parsed []dagger.BuildArg
	for _, arg := range *argss {
		s := strings.Split(arg, "=")
		if len(s) != 2 {
			return parsed, fmt.Errorf("invalid argument: %s. Use the format NAME=VALUE.", s)
		}

		parsed = append(parsed, dagger.BuildArg{Name: s[0], Value: s[1]})
	}

	return parsed, nil
}
