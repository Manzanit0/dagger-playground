package build

import (
	"context"
	"fmt"
	"os"
	"strings"

	"dagger.io/dagger"
	"github.com/manzanit0/dagger-playground/pkg/aws"
)

type BuildAndPushOptions struct {
	AwsRegion            string
	AwsEcrRepositoryName string
	AwsEcrURI            string
	Local                bool
	GitRepositoryURL     string
	GitRepositoryBranch  string
	DockerfilePath       string
	// TODO: this is a programatic interface... might as well make it correctly typed, i.e. key/value pairs.
	BuildArgs []string
}

// BuildAndPush builds the dockerfile and pushes it to ECR.
// Note: opts.BuildArgs is of the form of []string{{"FOO=BAR", "BAZ=LAR"}}
func BuildAndPush(ctx context.Context, opts *BuildAndPushOptions) error {
	args, err := parseBuildArgs(opts.BuildArgs)
	if err != nil {
		return fmt.Errorf("parse build args: %w", err)
	}

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return fmt.Errorf("dagger connect: %w", err)
	}
	defer client.Close()

	awsClient, err := aws.NewClient(ctx, opts.AwsRegion)
	if err != nil {
		return fmt.Errorf("new AWS client: %w", err)
	}

	username, password, err := awsClient.GetECRUsernamePassword(ctx)
	if err != nil {
		return fmt.Errorf("get ECR credentials: %w", err)
	}

	exists, err := awsClient.EnsureRepositoryExists(ctx, opts.AwsEcrRepositoryName)
	if err != nil {
		return fmt.Errorf("ensure ECR repository exists: %w", err)
	}

	if exists {
		fmt.Println("Skipped ECR repository creation, already exists")
	} else {
		fmt.Println("Created ECR repository")
	}

	var workspace *dagger.Directory
	if opts.Local {
		workspace = prepareLocalWorkspace(client, opts.DockerfilePath)
	} else {
		workspace = prepareRemoteWorkspace(client, opts.GitRepositoryURL, opts.GitRepositoryBranch, opts.DockerfilePath)
	}

	ref, err := client.
		Container().
		WithRegistryAuth(
			opts.AwsEcrURI,
			username,
			client.SetSecret("registryPassword", password),
		).
		Build(workspace, dagger.ContainerBuildOpts{
			Dockerfile: "Dockerfile",
			BuildArgs:  args,
		}).
		Publish(ctx, opts.AwsEcrURI+"/"+opts.AwsEcrRepositoryName)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Published image to :%s\n", ref)
	return nil
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

func parseBuildArgs(argss []string) ([]dagger.BuildArg, error) {
	var parsed []dagger.BuildArg
	for _, arg := range argss {
		s := strings.Split(arg, "=")
		if len(s) != 2 {
			return parsed, fmt.Errorf("invalid argument: %s. Use the format NAME=VALUE.", s)
		}

		parsed = append(parsed, dagger.BuildArg{Name: s[0], Value: s[1]})
	}

	return parsed, nil
}
