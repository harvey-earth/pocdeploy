# Welcome to pocdeploy! 
pocdeploy is a Golang CLI tool to deploy a POC app (in this case django-polls) with a CloudNative PG backend to a Kubernetes cluster.
This tool can deploy locally to Kind with plans to extend to AWS and other cloud providers to create a fully functioning environment in a single command.
Currently implemented frameworks for the frontend are simple Ruby on Rails and Django apps.

## Table of Content
[Prerequisites](#prerequisites)  
[Getting Started](#getting-started)  
[How it Works](#how-it-works)  
[Roadmap](#roadmap)  

## Prerequisites
- Docker installed and socket available
    - Docker Desktop Mac - In Settings > Advanced > Allow the default Docker socket to be used
- Kind Installed (along with prereqs)
- Install frontend codebase to a directory
    - Default is to use `third_party/django-polls`
- Install patch files
    - Default is to use `build/patches/`
- Install Dockerfile.frontend
    - Default is to use `build/Dockerfile.frontend`

## Getting Started
1. Get/Make pocdeploy binary and download the [pocdeploy.yaml](https://github.com/harvey-earth/pocdeploy/blob/main/pocdeploy.yaml) file.
1. Fill out the required credentials/values in the pocdeploy.yaml file.
    - Set frontend framework
        - `django` or `ror`
    - Set `check_path` to the URL path of the health check (ex. "/health")
2. Run `pocdeploy create`.
3. Run `pocdeploy delete` when done to clean up resources.

## How it Works
The command starts by standing up a Kubernetes cluster specified by the `--type` flag (Only Kind is fully implemented).
The code within the frontend.path variable will be patched with any patch files in the frontend.patch_dir directory.
For Django, the requirements.txt file is copied to the frontend code directory if it doesn't exist.
Next the Dockerfile at frontend.dockerfile will be used to create an image with the name from frontend.image and version frontend.version.
When the cluster is ready, the frontend application is deployed along with CloudNative PG as a backend.


The tool can deploy Django and Ruby on Rails frameworks that use Postgresql backends.

## Roadmap

- add Zap logger and verbose, debug flags
- `update` command that will build a new image, upload it to the cluster, and update the deployment to use the new image
- EKS cluster
- AKS cluster
- GKE cluster
- `install` command to install a default pocdeploy.yaml file from the binary
