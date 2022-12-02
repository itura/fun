# build-a-build

Build pipeline automation which generates GitHub Actions workflow definitions to run itself.

See the [pipeline config](/pipeline.yaml) input and the [workflow definitions](/.github/workflows/ci-cd.yaml) output.

## Usage
```
# from project root
go run tmty/build-a-build ./pipeline.yaml generate ./.github/workflows/ci-cd.yaml
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
- Application: Helm deployment
  - can be dependent on Artifacts or other Applications
  - can have Helm values injected from the GitHub secret or environment contexts
  - `deploy-application` command
    1. runs `helm update --install`, setting values defined in config as well as `repo` and `tag` values for Artifact images.
- Pipeline: set of Artifact and Application definitions
    - change detection
      - commands for each Artifact and Application will only take action if associated paths have changed since the previous commit
      - current limitation/bug is that when pushing multiple commits at once, change detection will only operate on the most recent 2 commits 
    - commands for both Artifacts and Applications will run in parallel based on dependencies

## Prerequisites
- GKE cluster
- Key for SA with permissions:
  - administrate k8s resources
  - push to container repository
- Workload identity for SA