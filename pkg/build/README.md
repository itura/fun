# build-a-build

Build pipeline automation which generates GitHub Actions workflow definitions to run itself.

See the example [pipeline config](./example/pipeline.yaml) input and the [workflow definitions](./example/workflow.yaml) output.

## Usage
```
# generate example
go run ./cmd/build generate pkg/build/example/workflow.yaml --config pkg/build/example/pipeline.yaml

# use in other repos
go run github.com/itura/fun/cmd/build generate ./.github/workflows/ci-cd.yaml
```

## Concepts

- Artifact: set of Docker images defined by multiple targets in a single Dockerfile
  - "verify" image used to test & lint source code
  - "app" image which runs the process defined by the source code
  - `build-artifact` command
    1. builds verify image
    2. runs verify image (terminate if failure)
    3. builds app image 
    4. tags and pushes app image
- Application: Helm chart or Terraform config
  - can be dependent on Artifacts or other Applications
  - `deploy-application` command either:
    1. runs `helm update --install`, setting values defined in config as well as `repo` and `tag` values for Artifact images.
    2. runs `terraform apply`
- Pipeline: set of Artifact and Application definitions
    - commands for both Artifacts and Applications will run in parallel based on dependencies using GH Actions job dependencies

## Features

### Helm value injection
For an Application, you can specify values to be set for the `helm upgrade`. This can take 3 forms:

```yaml
# pipeline.yaml
applications:
  - id: api
    path: helm/app
    values:
    # static value
    - key: app.name
      value: cool-api
    # load value from GH Actions `env`. 
    - key: app.lifecycle
      envValue: LIFECYCLE
    # load value from GH Actions `secret`
    - key: postgresql.auth.postgresPassword
      secretValue: DB_PASSWORD
```
Environment variables and secrets referenced by values must be populated separately.

### Change detection
Commands for each Artifact and Application will only perform their full actions if associated paths have changed since the previous commit.

If an Artifact's source has not changed, the image will not be built, but a new tag for the current commit will be added to the previous image. This makes it so that Applications can use the same tag for all Artifacts in that build.

A current limitation/bug is that when pushing multiple commits at once, change detection will only operate on the most recent 2 commits .

## Prerequisites
- GCP
- Workload identity for SA [link](https://github.com/google-github-actions/auth#setting-up-workload-identity-federation)
- Artifact Registry API enabled
- GKE cluster
- GitHub Actions
  - envs:
    - PROJECT_ID
    - WORKLOAD_IDENTITY_PROVIDER
    - SERVICE_ACCOUNT