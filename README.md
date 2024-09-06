# Welcome to pocdeploy! 
For this assessment, I decided to create a deployment tool to deploy the environment in a single command.
This creates a paved road for developers so they can make changes and spin up a new cluster with that code.
This shouldn't require a long list of steps to follow, as this introduces errors and time wasted.
This tool can be extended to work for other frontend frameworks and K8s cluster locations.

pocdeploy is a Golang CLI tool to deploy a POC app (in this case django-polls) with a CloudNative PG backend to a Kubernetes cluster.
This tool can deploy locally to Kind with plans to extend to AWS and other cloud providers to create a fully functioning environment in a single command.

I'll admit upfront that I had much higher expectations for this.
I was planning on additionally deploying to EKS, AKS, and GKE.
I had two more commands planned: `validate` to ensure all Kubernetes components were running correctly, and `update` which would build a new Docker image and upload it to the cluster.

## Table of Content
[Getting Started](#getting-started)
[How it Works](#how-it-works)

## Prerequisites
- Docker installed and socket available
    - Docker Desktop Mac - In Settings > Advanced > Allow the default Docker socket to be used
- Kind Installed (along with prereqs)
- Install frontend codebase
    - Default is to use `third_party/django-polls`
- Install patch files
    - Default is to use `build/patches/`
- Install Dockerfile.frontend
    - Default is to use `build/Dockerfile.frontend`

## Getting Started
1. Fill out the required credentials/values in the pocdeploy.yaml file.
2. Run `pocdeploy create`.
4. Run `pocdeploy delete` when done to clean up resources.

## How it Works
The command starts by standing up a Kubernetes cluster specified by the `--type` flag (Only Kind is fully implemented).
The code within the frontend.path variable will be patched with any patch files in the frontend.patch_dir directory.
The requirements.txt file is copied to the frontend code directory if it doesn't exist.
Next the Dockerfile at frontend.dockerfile will be used to create an image with the name from frontend.image and version frontend.version.
When the cluster is ready, the frontend application is deployed along with CloudNative PG as a backend.

## Design Decisions

### Patching
There are some changes in the Django Polls App to run the application in the manner specified.
I chose to use patch files so that any updates to the Django Polls code could be included in the future easily.
This also allows other apps that may need changes to the code to be deployed by this tool.
If the patches directory is empty then patching can be silently skipped.
There are some minor changes to make the application more production worthy (turn off debug, set secret key).

### Building Image
Other Dockerfiles can be used to add to extensibility.

### Cluster
I separated the Terraform code for each provider into separate directories.
In the case of Kind, the configuration is done with a single file.
These are embedded into the golang package so the user can run the binary from anywhere.

### Kubernetes
The CloudNative PG operator is installed and used for deploying the backend.
The frontend is deployed with a Deployment, Service, and Ingress.
The Service is of type LoadBalancer for cloud clusters, and NodePort for Kind.

### Misc
I generated a requirements.txt file using `pipreqs` and `pip-tools` with `pipreqs --savepath=requirements.in && pip-compile` within the Django Polls app directory.
