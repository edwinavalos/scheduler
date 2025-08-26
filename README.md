# scheduler

This is a golang program that schedules a set of services based on precomposed service definitions defined in golang code.

For the moment this is going to be a monolith that receives API requests, and does the deployment of services to a local containerd installation.

Load balancing will be taken care of by a independent loadbalancer that will direct traffic to this API to be serviced by the containerd containers. The scheduler program will be a standalone golang binary running on the baremetal server that is the host for the containerd containers.

## GRPC

This project uses GRPC as the primary communication method to the server. As such there are proto files contained in the `proto` folder that define our container scheduling service.

The service should allow for the user to specify a specification of containers that make up an application stack. The stack will primarily consist of a frontend container, a backend container, and a postgresql container.

## Cobra CLI & Viper

Cobra is used as the general command line interface of the application, we launch the scheduling service by runnning `./scheduler run. The host and port of the server should be defaulted to localhost:8000.

THere should be a configuration file that contains all configurable values for the application server.

## Services

EnvironmentService is the service that has CRUD operations for Environments.

Environments are specified via an EnvironmentSpecification.

An EnvironmentSpecification is the specification for the configurable options for running  containerD containers.
