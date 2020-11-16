 > One repository to rule them all.  
 
# Summary

By a monorepo I mean a single repository that contains multiple projects written in different languages as well. 
It's a proposal on multi-stack monorepo which is based on root directory stack segregation and uses shared domain models in the context of the same language to make code reusable.

# Monorepo structure

## Separate service folders
Services are split into separate folders, as for example Go services are located in /app/go and Python services in /app/Python while keeping the root directory clean. 

Domain models are stored in the domain directory while they are still accessible for monorepo services.

## CMD pattern
It's a common pattern in Go to store `main.go` files in `cmd` directory. This pattern cleans up the root service directory from `main.go` files and makes it easier to handle several main files if it's needed.

## Running the app locally

## Running the Docker containerized application
Get in the required service directory and execute the `docker-compose up --build` command to create and start the containers. 

## Principles and tools used to work on the repo (Go directory for now)
Integrated pgAdmin, PostgreSQL, NSQ (Admin, lookup service, queue) for easy access to admin tools out of the box.

CQRS - Command Query Responsibility Segregation, commands and queries are split into separate pieces:
- command: https://github.com/VladimirAndrianov96/monorepo-example/blob/main/app/Go/domain/models/user/command.go
- query: https://github.com/VladimirAndrianov96/monorepo-example/blob/main/app/Go/domain/models/user/query.go

DDD - Domain-Driven Design
- shared domain logic: https://github.com/VladimirAndrianov96/monorepo-example/tree/main/app/Go/domain

API and RPC for service communication.

Ginkgo, Gomega BDD frameworks to help keep tests readable and simple, BDD test scenarios are much more closer to the actual use-cases.

Gorm as Go ORM to make migrations simple.

YAML file format to store static configuration details, ENV variables can be used as an alternative to static details.


Mock tests to test communication with external network components locally.
Docker and Docker-Compose to wrap the stuff and ease the deployment.
SSL certificates for services available from the web.


## TODO
- Make configuration more flexible (switch between docker and local)
- Add session handling endpoints
- Integrate centralized logging solution
- Add front-end dashboard to visualize the thing
- Move from NSQ to RabbitMQ or create `jobs` db in table to fix the possible inconsistency
- Add Kubernetes manifests
- Improve Python and test service structure
- Add more fields to tables to provide a better demo on indexes
- Add NGINX as proxy for the future front-end
- Improve shutdown logic, use channels
