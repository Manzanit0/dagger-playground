# 🏖 Dagger.io playground

A scratchpad playing with https://dagger.io

## Pipelines

Some pipelines I'm thinking might be interesting to try out:

### Release

- [ ] Bump the tag of a private git repository

Thoughts: Dagger didn't really help here. While it does provide a super-clean abstraction for querying git repositories, it doesn't currently allow for writing to them, so ended up reaching to other 3rd party libs.

### Build-Push to ECR

- [x] Clone remote private repository
- [x] Build Docker image
- [x] Create AWS ECR repository if it doesn't exist
- [x] Push image to AWS ECR

Thoughs: I think this is the ideal use-case for Dagger. A simple query to a git repository and building a container. It was a breeze.

### Deploy

- [ ] Open PR to GitHub repository with an update to a YAML file (hello Flux)

### Notify

- [ ] Send a Slack notification with deployment trigger

## Application

- [x] CLI that applies the pipeline
- [ ] Service that listens to GH webhooks and applies the pipelines
- [ ] GitHub Action that runs the pipelines ([example](https://docs.dagger.io/620941/github-google-cloud/))

## Thoughts on Dagger modules

I started tinkering with [Dagger modules](https://docs.dagger.io/zenith/developer/go/525021/quickstart), but came across some expected friction when dealing with authentication to providers such as AWS: when you run `dagger call foo`, that wraps the function call in a container which means that the credentials available in the host are no longer available.

A workaround could be to create a set of client id and secret for the dagger module, but this generally isn't the desired outcome since in AWS-land assuming roles tends to be preferred. If that's the case, then not using modules in favour of `dagger run` becomes simpler. We just miss out on the reusable modularity of the "Daggerverse".
