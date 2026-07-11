# ChaosGuard

> Automated Chaos Engineering Platform for Docker-based Microservices

ChaosGuard is an open-source chaos engineering platform built with Go to help developers test the resilience of Docker-based microservices by intentionally injecting failures in a controlled environment.

The goal of this project is simple: instead of waiting for failures to happen in production, simulate them during development and observe how the application behaves.

ChaosGuard is currently under active development and is being built incrementally with a focus on clean architecture, modular design, and production-oriented engineering practices.

---

## Why ChaosGuard?

Modern applications are built using multiple services that depend on each other. A failure in one service can quickly affect the entire system if proper recovery mechanisms are not in place.

ChaosGuard aims to help developers answer questions like:

* What happens if the payment service suddenly stops?
* Does the application recover after a container restart?
* Are failures handled gracefully?
* Are there enough metrics to understand what happened?
* How resilient is the application overall?

Instead of discovering these problems in production, ChaosGuard helps uncover them during development.

---

## Current Features

* Docker container discovery
* Chaos experiment scheduler
* Pause, Stop, Restart and Kill attacks
* Automatic recovery manager
* Safe mode to avoid attacking critical containers
* Prometheus metrics integration
* SQLite experiment persistence
* Configurable through YAML and environment variables
* Structured logging
* Unit tests across core modules

---

## Project Structure

```text
ChaosGuard
├── cmd/                # CLI entry point
├── internal/
│   ├── domain/         # Domain models and interfaces
│   ├── infra/          # Docker and SQLite implementations
│   └── usecase/        # Scheduler, attacks, recovery
├── pkg/                # Shared packages
├── configs/
└── docs/
```

The project follows a Clean Architecture approach where business logic remains independent from infrastructure.

---

## Technology Stack

* Go
* Docker SDK
* Cobra CLI
* SQLite
* Prometheus
* Zerolog
* Viper

Planned additions:

* React Dashboard
* Grafana
* REST API
* Report Generation
* Issue Detection Engine

---

## Getting Started

Clone the repository

```bash
git clone https://github.com/vaishnavanS/ChaosGuard.git
cd ChaosGuard
```

Install dependencies

```bash
go mod tidy
```

Run the project

```bash
go run cmd/chaosguard/main.go
```

Run tests

```bash
go test ./...
```

---

## Roadmap

### Completed

* CLI foundation
* Docker integration
* Scheduler
* Attack engine
* Recovery manager
* Metrics collection
* SQLite persistence

### In Progress

* Runtime wiring
* Experiment lifecycle
* REST API

### Planned

* Web dashboard
* Grafana dashboards
* Rule-based issue detection
* Recommendation engine
* Report generation
* Additional chaos experiments

---

## Project Status

ChaosGuard is currently in active development.

The core backend architecture has been implemented, and the focus is now shifting toward runtime integration, APIs, dashboards, and resilience analysis.

---

## Why I Built This

I started this project to better understand how modern distributed applications behave when individual services fail.

Rather than building another CRUD application, I wanted to work on something that combines backend development, Docker, observability, testing, and software reliability into a single project.

ChaosGuard is also an opportunity for me to learn more about Site Reliability Engineering (SRE), DevOps practices, and chaos engineering while building a real open-source tool.

---

## Contributing

Contributions, bug reports, and suggestions are always welcome.

If you find an issue or have an idea that could improve ChaosGuard, feel free to open an issue or submit a pull request.

---

## License

This project is licensed under the MIT License.
