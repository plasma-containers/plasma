# plasma 🌌

[Plasma Changelog](CHANGELOG.md)

## What is it?

Plasma is an application that when deployed to Docker host, allows to deploy Compose projects
using its own HTTP API, with using the same Compose file you deploy your project locally.

## What does it do?

### plasma-server

- Deploy it locally or on hobby VPS.
- It exposes a HTTP api to upload Compose files.  
- Manages Docker containers and volumes created through it.  
- Uses healthchecks to check if containers are healthy.  
- If not, kills them and redeploys them.  
- Will feature JWT authentication to be secure.  
- Will allow to fetch container logs, remove a project, restart container etc.  

### plasma - CLI

- Interacts with plasma-server API.
- Pushes Compose files to plasma-server.
- Can check all managed resources' state.
- Can deploy plasma-server locally for testing (see Quickstart).

## Quickstart

Run
```sh
plasma serve 
```
to run plasma-server in local docker.  
Run it without arguments to see full usage, how to push Compose file to it etc.  
