# Getting Started

## Prerequisites

- Go 1.25+
- Docker / Docker Compose

## Start Neo4j

```bash
docker compose up -d
```

- **Browser UI**: http://localhost:7474 (login: `neo4j` / `password`)
- **Bolt endpoint**: `bolt://localhost:7687` (used by the Go app)

## Seed Data

```bash
make seed
```

This populates Neo4j with AI supply chain companies and their relationships, including NVIDIA, TSMC, OpenAI, xAI, Mistral AI, Cohere, Broadcom, and more.

## Run the App

```bash
make run
```

The server starts on http://localhost:8080.

## Try It Out

```bash
# Create a company
curl -X POST localhost:8080/api/v1/companies -H 'Content-Type: application/json' \
  -d '{"name":"Perplexity","type":"ai_lab","founded":2022,"hq":"San Francisco"}'

# Create a relationship
curl -X POST localhost:8080/api/v1/relationships/supplies-chips-to -H 'Content-Type: application/json' \
  -d '{"supplier_name":"NVIDIA","client_name":"Perplexity"}'

# Get company with all relationships
curl localhost:8080/api/v1/companies/NVIDIA

# List all companies
curl localhost:8080/api/v1/companies
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/companies` | Create a company |
| GET | `/api/v1/companies` | List all companies |
| GET | `/api/v1/companies/:name` | Get company with relationships |
| POST | `/api/v1/relationships/supplies-equipment-to` | ASML supplies equipment to TSMC |
| POST | `/api/v1/relationships/manufactures-for` | TSMC manufactures for NVIDIA |
| POST | `/api/v1/relationships/supplies-chips-to` | NVIDIA supplies chips to OpenAI |
| POST | `/api/v1/relationships/provides-cloud-for` | AWS provides cloud for Anthropic |
| POST | `/api/v1/relationships/competes-with` | OpenAI competes with Anthropic |
