# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
make build          # Build binary to bin/learn-neo4j
make run            # Run the server
make test           # Run all tests with verbose output
make seed           # Seed sample data into Neo4j
make clean          # Clean test cache and bin/
```

Requires a running Neo4j instance. Start one with:
```bash
docker compose up -d    # Neo4j on bolt://localhost:7687, browser on :7474
```

Default credentials: `neo4j/password` (configured in `application.yml`).

## Architecture

Go REST API using Gin, backed by Neo4j graph database. Layered architecture:

- **`cmd/learn-neo4j/`** - Application entrypoint, config loading (via `adder` library + `application.yml`), logger setup
- **`cmd/seed-data/`** - Standalone script to populate Neo4j with sample movie data
- **`internal/api/http/`** - HTTP layer: router, handlers, DTOs, middleware (request logging, error handling)
- **`internal/service/`** - Business logic layer
- **`internal/repository/`** - Neo4j Cypher queries via `neo4j-go-driver/v5`
- **`internal/domain/`** - Domain models (Movie, Person, ActedIn, MovieWithCast)
- **`pkg/neo4j/`** - Neo4j client wrapper (connection setup/teardown)

All repository methods use `neo4j.ExecuteQuery` with `EagerResultTransformer` and explicitly target the `"neo4j"` database.

Config is loaded from `application.yml` with env var override support (dot-separated keys become underscored env vars, e.g. `NEO4J_URI`).

## API Routes

All CRUD endpoints are under `/api/v1` (movies, persons, acted-in relationships). Health check at `/health`.
