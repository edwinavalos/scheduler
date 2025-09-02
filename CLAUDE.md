# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based container scheduler that manages containerD containers through a gRPC API. The scheduler is designed as a monolithic service that receives API requests and deploys services to a local containerD installation.

## Architecture

- **Primary Communication**: gRPC API as the main interface
- **CLI Framework**: Cobra CLI with Viper for configuration management
- **Container Runtime**: containerD for container orchestration
- **Load Balancing**: Handled by an independent load balancer that directs traffic to the API
- **Deployment Target**: Standalone Go binary running on bare metal server hosting containerD containers

## Key Components

- **Proto Files**: Service definitions located in `proto/` folder defining the container scheduling service
- **EnvironmentService**: Core service providing CRUD operations for Environments
- **EnvironmentSpecification**: Configuration specification for containerD container options
- **Application Stacks**: Typically consist of frontend container, backend container, and PostgreSQL container

## Development Commands

### Makefile Targets (Recommended)
- **Build**: `make build` - Build the scheduler binary
- **Run**: `make run` - Build and start the server (localhost:8000)
- **Clean**: `make clean` - Remove build artifacts
- **Help**: `make help` - Show all available targets
- **Dependencies**: `make deps` - Install/update Go dependencies
- **Test**: `make test` - Run all tests
- **Format**: `make fmt` - Format Go code
- **Proto Generation**: `make proto` - Generate protobuf code (when proto files exist)

### Direct Go Commands
- **Install Dependencies**: `go mod tidy`
- **Build**: `go build -o scheduler`
- **Start Server**: `./scheduler run` (defaults to localhost:8000)
- **Help**: `./scheduler --help` or `./scheduler run --help`
- **Custom Config**: `./scheduler run --config /path/to/config.yaml`
- **Custom Host/Port**: `./scheduler run --host 0.0.0.0 --port 9000`

## Configuration

- Default config file: `.scheduler.yaml` (in current directory or home directory)
- Configuration managed by Viper with support for YAML files, environment variables, and command-line flags
- Host and port default to localhost:8000
- Environment variables: prefix with `SCHEDULER_` (e.g., `SCHEDULER_HOST`, `SCHEDULER_PORT`)

## Project Structure

- `main.go`: Entry point
- `cmd/`: Cobra CLI commands
  - `cmd/root.go`: Root command and configuration setup
  - `cmd/run.go`: Server start command with gRPC server skeleton
- `.scheduler.yaml`: Default configuration file