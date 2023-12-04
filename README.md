# 🏖 Dagger.io playground

A scratchpad playing with https://dagger.io

## Pipelines

Some pipelines I'm thinking might be interesting to try out:

### Release

- [ ] Bump the tag of a private git repository

### Build-Push to ECR

- [x] Clone remote private repository
- [x] Build Docker image
- [x] Create AWS ECR repository if it doesn't exist
- [x] Push image to AWS ECR

### Deploy

- [ ] Open PR to GitHub repository with an update to a YAML file (hello Flux)

## Application

- [x] CLI that applies the pipeline
- [ ] Service that listens to GH webhooks and applies the pipelines
