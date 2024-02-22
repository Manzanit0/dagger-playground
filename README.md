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

```sh
dagger run go run ./cmd/build \
    --git-repository git@github.com:Manzanit0/weatherwarnbot.git \
    --git-repository-branch master \
    --dockerfile Dockerfile \
    --aws-ecr-uri=<uri here>
```

### Deploy

- [ ] Open PR to GitHub repository with an update to a YAML file (hello Flux)

### Notify

- [ ] Send a Slack notification with deployment trigger

## Application

- [x] CLI that applies the pipeline
- [x] Service that listens to GH webhooks and applies the pipelines
- [x] GitHub Action that runs the pipelines ([example](https://docs.dagger.io/620941/github-google-cloud/))

### Running pipelines via GH webhooks

It was actually pleasantly easy to set this up. I created a sample service under `cmd/webhook`. GitHub webhooks give you all the information you need not just to run the build&push pipeline, but also to chain a Slack command or something similar with a diff of the code and other related information.

I'm curious how it will work once the service is dockerised; if running the dagger client will pose some challenges or not. My guess is that it should be fine because when you run `dagger run` it's doing exactly that: running the client inside a container (the dagger engine).

In terms of scaling that in an organisation, I think the challenge is setting up the webhook the right way. I can see two options here: either [an organisation webhook](https://docs.github.com/en/webhooks/types-of-webhooks#organization-webhooks) or a [GitHub App webhook](https://docs.github.com/en/webhooks/types-of-webhooks#github-app-webhooks). I
not really familiar with the constraints of the GitHub App webhooks, but it does sound like that would be the ideal approach to not get undesired events, i.e. repositories that you don't want to build. I also wonder what would be the right way to set up that application so that it's dead-simple to get it working. The moment you have to start dealing with approval processes and the like to install the app, it starts sucking.

This approach seems overall the best one to scale, as opposed to GitHub actions. The main reason being the ease of rollouts and configuration by teams.

### Using GitHub Actions to wrap Dagger

From an initial stab, it's not bad. It allows the designated DevOps person or team to write most of the stuff in Go and then just do some simple wrapping in YAML. However, there are some bits which I'd still like to look into such as looking into tackling the OIDC workflow with Go too so we don't have to leverage the `aws-actions/configure-aws-credentials` actions. This would reduce the surface area of the dependency on GHA even more.

I'd also need to think about what's the best way of versioning, how to roll those out, etc. I have a hunch this is just a good 'ol GHA versioning problem though.

## Thoughts on Dagger modules

I started tinkering with [Dagger modules](https://docs.dagger.io/zenith/developer/go/525021/quickstart), but came across some expected friction when dealing with authentication to providers such as AWS: when you run `dagger call foo`, that wraps the function call in a container which means that the credentials available in the host are no longer available.

A workaround could be to create a set of client id and secret for the dagger module, but this generally isn't the desired outcome since in AWS-land assuming roles tends to be preferred. If that's the case, then not using modules in favour of `dagger run` becomes simpler. We just miss out on the reusable modularity of the "Daggerverse".
