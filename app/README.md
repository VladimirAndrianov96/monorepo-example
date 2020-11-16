# App directory

App directory is meant to store the actual application code there, while rest of the stuff can be placed in the root monorepo directory.

As for example, there is two Golang services with HTTP API cross-communication
- Users API
- Test API for communication testing

and Python service which consumes messages coming from Users API service.

Docker-compose and Kubernetes configuration files can be stored here as well. 
