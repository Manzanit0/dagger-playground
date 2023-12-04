package main

import (
	"context"
)

type RegistryInfo struct {
	uri      string
	username string
	password string
}

// InitRegistry creates and/or authenticate with an ECR repository
func InitRegistry(ctx context.Context, awsClient *AWSClient, repoURI string) *RegistryInfo {
	username, password, err := awsClient.GetECRUsernamePassword(ctx)
	if err != nil {
		panic(err)
	}

	return &RegistryInfo{repoURI, username, password}
}
