# SAGA Pattern implemented in Go

This is an example of how a [SAGA Pattern](https://medium.com/cloud-native-daily/microservices-patterns-part-04-saga-pattern-a7f85d8d4aa3) can be implemented using an Choreography structure in Go. This is purely educational.

This project will also showcase (using different branches) how we can simplify the life of the developer experience with the introduction of tools such as Skaffold, Helm and Make.

# Structure of the project
For this project, we are going to show a minimal setup of 3 microservices:
- Order: Service incharge of handling all the orders that are made to our restaurant
- Inventory: Service incharge of handling all the deliveres to the user

# How to run it?
> [!IMPORTANT]
> On every branch you can find the SAGA pattern implemented the same way, the only thing that will change is our toolset

This is a complex question as you may think, but this are the step depending on the branch that you are placed:

## Basic tooling (branch `barebones-approach`)
For this setup our main tools are:
- Docker compose for development

And the idea to put this into a production stage with multiple images, looking to deploy in something similar to ECS.

```bash
# Run our application without cleaning the databases
bash setup/run-app.sh

# Run our application cleaning the database (cleaning volumes)
docker setup/run-app.sh -c / --clean
```

## Easier Developer Experience (branch `easier-dev-xp`)
In this stage, we are introducing:
- 'Skaffold' for easier local development to avopid using docker compose for local development
- 'Make' for building our application in a more automated way
- 'K8s' (using minikube), to a have a better control over our containers

>[!IMPORTANT]
> You must run the following command before making any docker related stuff 

```
skaffold config set --global local-cluster true eval $(minikube -p custom docker-env)
```
To build and deploy just run
```
make dev
```


## Don't want to repeat, let's template

In this stage, Helm is introduce to avoid duplication inside of our YAML's and have a versioning of our deployments.

# Refernces
This were some of the posts and articles that I read to make this project:
- [Database per Microservice pattern](https://microservices.io/patterns/data/database-per-service.html)