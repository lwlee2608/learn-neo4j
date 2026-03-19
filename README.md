# learn-neo4j

Go REST API backed by Neo4j, modelling the AI supply chain as companies and company-to-company relationships.

## Graph Schema

The AI-assisted query and graph expansion flows share one graph schema definition in `internal/graphschema/schema.go`.

- `internal/graphschema` is the single source of truth for allowed labels, relationship types, properties, and the schema prompt shown to the LLM.
- `internal/nlquery` uses that schema to constrain read-only Cypher generation and validation.
- `internal/graphexpand` uses the same relationship allowlist to validate graph write plans before applying them.

When the graph model changes, update `internal/graphschema/schema.go` so both AI paths stay aligned.

## Commands

### `cmd/learn-neo4j`

Main API server. Starts an HTTP server (Gin) exposing REST endpoints for querying and creating companies and their relationships in the graph database.

### `cmd/seed-data`

Populates Neo4j with sample AI supply chain data — companies (NVIDIA, TSMC, OpenAI, xAI, Mistral AI, Cohere, Broadcom, etc.) and relationships between them (SUPPLIES_CHIPS_TO, MANUFACTURES_FOR, PROVIDES_CLOUD_FOR, etc.). Clears existing data before seeding.

### `cmd/ask-cypher`

Natural language query tool. Takes a plain-English question, uses an LLM (via OpenRouter) to translate it into a Cypher query, executes it against Neo4j, and returns the answer.

### `cmd/expand-graph`

Graph expansion tool. Given a keyword (e.g. a company name), uses an LLM and web search (Exa AI) to discover new entities and relationships, then writes them into the Neo4j graph.

## Docs

- `docs/getting-started.md` - local setup, seeding, and API examples
- `docs/ai-graph-schema.md` - how the shared graph schema drives LLM prompts and validation
- `docs/graph-rag-roadmap.md` - how the current repo fits GraphRAG and what to improve next
- `docs/forecasting/README.md` - phased plan for stock-price forecasting integration
- `docs/cypher-basics.md` - Cypher language primer
- `docs/neo4j-browser.md` - using Neo4j Browser locally
- `docs/neo4j-concepts.md` - graph database concepts used in this project
