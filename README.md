# learn-neo4j

Go REST API backed by Neo4j, modelling the AI supply chain (companies, chips, and their relationships).

## Commands

### `cmd/learn-neo4j`

Main API server. Starts an HTTP server (Gin) exposing REST endpoints for querying and creating companies, chips, and relationships in the graph database.

### `cmd/seed-data`

Populates Neo4j with sample AI supply chain data — companies (NVIDIA, TSMC, OpenAI, etc.) and relationships between them (SUPPLIES_CHIPS_TO, MANUFACTURES_FOR, PROVIDES_CLOUD_FOR, etc.). Clears existing data before seeding.

### `cmd/ask-cypher`

Natural language query tool. Takes a plain-English question, uses an LLM (via OpenRouter) to translate it into a Cypher query, executes it against Neo4j, and returns the answer.

### `cmd/expand-graph`

Graph expansion tool. Given a keyword (e.g. a company name), uses an LLM and web search (Exa AI) to discover new entities and relationships, then writes them into the Neo4j graph.
