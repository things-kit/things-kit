# Things-Kit Project Overview

## Purpose
Things-Kit is a modular, opinionated microservice framework for Go, built on Uber Fx. It aims to bring the productivity and developer experience of frameworks like Spring Boot to the Go ecosystem.

## Vision
Create a "batteries-included, but optional" ecosystem where developers can build robust, production-ready services by simply composing pre-built, independent modules. This avoids the boilerplate of wiring up common components while retaining Go's idiomatic simplicity.

## Core Philosophy
1. **Modularity First**: Every piece of infrastructure (gRPC, Kafka, Redis, etc.) is a self-contained, versionable Go module
2. **Dependency Injection is King**: Built entirely on Uber Fx for application lifecycle and dependency management
3. **Program to an Interface**: Core components are defined by abstractions (interfaces), not concrete implementations
4. **Convention over Configuration**: Sensible defaults with full override capability
5. **Lifecycle Aware**: All modules integrate with Fx lifecycle for graceful startup and shutdown
6. **Developer Experience**: Minimal boilerplate through generic helpers

## Tech Stack
- **Language**: Go 1.21+
- **DI Framework**: Uber Fx (go.uber.org/fx)
- **Configuration**: Viper (github.com/spf13/viper)
- **Logging**: Zap (go.uber.org/zap) - default implementation
- **gRPC**: google.golang.org/grpc
- **HTTP**: Gin (github.com/gin-gonic/gin)
- **Database**: database/sql with PostgreSQL
- **Redis**: github.com/redis/go-redis/v9
- **Kafka**: github.com/segmentio/kafka-go

## Target Users
Developers building Go microservices who want:
- Quick service bootstrapping
- Production-ready defaults
- Clean architecture through dependency injection
- Easy testing
- Pluggable components
