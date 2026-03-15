# Shared AI Graph Schema

The repository keeps one shared graph schema definition for AI-powered query generation and graph expansion.

## Source of Truth

The canonical schema lives in `internal/graphschema/schema.go`.

It defines:

- allowed node labels
- allowed relationship types
- allowed properties per label
- the prompt template used to describe the graph to the LLM

## Where It Is Used

### `internal/nlquery`

The natural-language query flow uses the shared schema to:

- build the graph description embedded into LLM prompts
- validate generated Cypher labels, relationship types, and properties
- keep read-only query generation constrained to the known graph shape

### `internal/graphexpand`

The graph expansion flow uses the same schema to:

- validate relationship types in generated expansion plans
- reject writes that reference unsupported graph edges

## Why This Exists

Before the refactor, schema details were duplicated across packages. That made it easy for the query path and write path to drift apart as the graph model changed.

With the shared package:

- schema updates happen in one place
- prompt generation and validation stay aligned
- new labels or relationships can be introduced more safely

## Updating the Graph Model

When you add or rename labels, relationship types, or allowed properties:

1. update `internal/graphschema/schema.go`
2. update repository/domain code that reads or writes the new graph shape
3. run `make test`

`pkg/ai/schema.go` serves a different purpose: it generates JSON Schema for Go structs used as LLM tool inputs. It does not inspect Neo4j directly and is not the source of truth for the graph model.
