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

## Run the App

```bash
make run
```

The server starts on http://localhost:8080.

## Try It Out

```bash
# Create a movie
curl -X POST localhost:8080/api/v1/movies -H 'Content-Type: application/json' \
  -d '{"title":"The Matrix","released":1999,"tagline":"Welcome to the Real World"}'

# Create a person
curl -X POST localhost:8080/api/v1/persons -H 'Content-Type: application/json' \
  -d '{"name":"Keanu Reeves","born":1964}'

# Create a relationship
curl -X POST localhost:8080/api/v1/acted-in -H 'Content-Type: application/json' \
  -d '{"person_name":"Keanu Reeves","movie_title":"The Matrix","role":"Neo"}'

# Get movie with cast
curl localhost:8080/api/v1/movies/The%20Matrix

# List all movies
curl localhost:8080/api/v1/movies

# List all persons
curl localhost:8080/api/v1/persons
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/movies` | Create a movie |
| GET | `/api/v1/movies` | List all movies |
| GET | `/api/v1/movies/:title` | Get movie with cast |
| POST | `/api/v1/persons` | Create a person |
| GET | `/api/v1/persons` | List all persons |
| POST | `/api/v1/acted-in` | Create an ACTED_IN relationship |
