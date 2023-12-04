package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
)

type AWSClient struct {
	region string
	cEcr   *ecr.Client
}

func NewAWSClient(ctx context.Context, region string) (*AWSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	cfg.Region = region
	client := &AWSClient{
		region: region,
		cEcr:   ecr.NewFromConfig(cfg),
	}

	return client, nil
}

func (c *AWSClient) GetECRAuthorizationToken(ctx context.Context) (string, error) {
	log.Printf("ECR GetAuthorizationToken for region %q", c.region)
	out, err := c.cEcr.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", err
	}

	if len(out.AuthorizationData) < 1 {
		return "", fmt.Errorf("GetECRAuthorizationToken returned empty AuthorizationData")
	}

	authToken := *out.AuthorizationData[0].AuthorizationToken
	return authToken, nil
}

// GetECRUsernamePassword fetches ECR auth token and converts it to username / password
func (c *AWSClient) GetECRUsernamePassword(ctx context.Context) (string, string, error) {
	token, err := c.GetECRAuthorizationToken(ctx)
	if err != nil {
		return "", "", err
	}

	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return "", "", err
	}

	split := strings.SplitN(string(decoded), ":", 2)
	if len(split) < 1 {
		return "", "", fmt.Errorf("invalid base64 decoded data")
	}

	return split[0], split[1], nil
}

func (c *AWSClient) EnsureRepositoryExists(ctx context.Context, name string) (bool, error) {
	list, err := c.cEcr.DescribeRepositories(ctx, &ecr.DescribeRepositoriesInput{RepositoryNames: []string{name}})
	if err != nil && !strings.Contains(err.Error(), "RepositoryNotFoundException") {
		return false, fmt.Errorf("create repository: %w", err)
	}

	if err == nil && len(list.Repositories) > 0 {
		return true, nil
	}

	_, err = c.cEcr.CreateRepository(ctx, &ecr.CreateRepositoryInput{
		RepositoryName:     &name,
		ImageTagMutability: types.ImageTagMutabilityMutable,
		Tags:               []types.Tag{{Key: s("created-by"), Value: s("dagger")}},
	})

	if err != nil {
		return false, fmt.Errorf("create repository: %w", err)
	}

	return true, nil
}

func s(ss string) *string {
	return &ss
}
