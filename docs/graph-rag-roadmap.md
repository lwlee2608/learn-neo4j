# GraphRAG and Knowledge Graph Roadmap

## Current State

This repository implements two AI-powered graph workflows:

- **NL-to-Cypher** (`cmd/ask-cypher`): an LLM agent translates natural language questions into schema-validated Cypher, executes against Neo4j, and synthesizes answers.
- **Graph Expansion** (`cmd/expand-graph`): an LLM agent searches the web via Exa AI, extracts entities and relationships, and MERGEs them into Neo4j.

This is a form of GraphRAG — the graph is the retrieval layer for LLM-driven question answering. More precisely it is **Text-to-Cypher** style GraphRAG, where retrieval is a generated graph query rather than vector similarity search.

What's missing for a full GraphRAG system: hybrid retrieval (graph + vector + documents), community-level summarization, and provenance tracking.

## Improvements

### Quick wins (high impact, small effort)

#### Few-shot examples in prompts

Add 3–5 example question→Cypher pairs to the agent system prompt template (`internal/nlquery/templates/agent_system_prompt.tmpl`). This is the single biggest lever for Cypher generation accuracy.

Good examples to include:

- simple lookup: "What type of company is TSMC?"
- relationship traversal: "Who supplies chips to OpenAI?"
- multi-hop: "Trace the supply chain from ASML to OpenAI"
- aggregation: "How many cloud providers does Anthropic use?"
- variable-length path: "Find all paths between ASML and OpenAI within 3 hops"

#### Error recovery in agent loop

When a Cypher query fails validation or execution, the `execute_cypher_query` tool returns a JSON error string. Coach the agent prompt to retry with a corrected query instead of giving up. Add explicit instructions like "If the tool returns an error, analyze what went wrong and try a different query."

#### Relationship properties

Add properties to edges in the schema: `since`, `value`, `notes`. This turns "TSMC manufactures for NVIDIA" into "TSMC manufactures for NVIDIA since 2016 (7nm and 5nm process nodes)". Update `internal/graphschema/schema.go` and the seed data.

#### Node description property

Add a `description` text property to Company nodes. This is useful on its own for richer answers, and is a prerequisite for vector search later.

### Graph model (medium effort)

#### More node types

The current graph has only `Company`. Adding more types enables richer queries:

| Node type    | Examples                      | Enables                                       |
| ------------ | ----------------------------- | --------------------------------------------- |
| `Product`    | H100, GPT-4, EUV lithography  | "what products does NVIDIA supply to OpenAI?" |
| `Technology` | CUDA, Triton, CoWoS packaging | "what technologies does TSMC use?"            |
| `Country`    | Taiwan, USA, Netherlands      | "what is the geographic concentration risk?"  |

Update `internal/graphschema/schema.go`, add corresponding relationship types, and extend the seed data.

#### Entity resolution and aliases

Store canonical names and aliases to prevent duplicates and improve match quality.

Examples: `TSMC` / `Taiwan Semiconductor Manufacturing Company`, `OpenAI` / `OpenAI, Inc.`

Options:

- an `aliases` array property on Company nodes, matched with `WHERE c.name = $name OR $name IN c.aliases`
- a separate `Alias` node type with `(:Alias)-[:ALIAS_OF]->(:Company)`

The second option is more flexible but adds query complexity.

### Retrieval (high effort, high impact)

#### Vector search (hybrid retrieval)

Store embeddings on nodes using Neo4j's native vector index. Before generating Cypher, do a vector similarity search to find relevant nodes, then scope the Cypher query to that subgraph.

Flow:

1. embed the user question
2. vector search to find top-k relevant nodes
3. feed those node names into the agent as context
4. agent generates Cypher scoped to the relevant subgraph

This handles fuzzy/semantic queries that pure Text-to-Cypher struggles with, like "companies at risk if there's a Taiwan earthquake."

#### Subgraph context injection

For complex questions, retrieve the 2-hop neighborhood around matched nodes, serialize as text, and include in the LLM prompt alongside Cypher results. This gives the LLM richer context for reasoning without needing a perfect Cypher query.

Add a `get_node_neighborhood` tool to the agent alongside `execute_cypher_query`. This helps with questions like "explain NVIDIA's role in the AI supply chain" where a single Cypher query can't capture the full picture.

#### Community summaries

Pre-compute text summaries for clusters in the graph (e.g., "the chip manufacturing cluster: ASML→TSMC→NVIDIA→OpenAI"). Use these as retrieval units for broad questions like "explain the AI chip supply chain." This is the core idea from Microsoft's GraphRAG paper.

### Knowledge graph quality

#### Provenance and confidence

Every graph fact should be traceable. Add metadata properties to nodes and relationships:

- `source_url` — where the fact came from
- `retrieved_at` — when it was extracted
- `confidence` — float score (1.0 for seed data, lower for web-extracted)

This makes the knowledge graph auditable and lets queries filter by confidence.

#### Time-aware relationships

Supply-chain facts change. Add temporal properties: `effective_from`, `effective_to`, `last_verified`.

This enables queries like "who currently manufactures for NVIDIA" and lets the system flag stale data for re-verification.

#### Safer graph expansion

`cmd/expand-graph` should be more conservative:

- require source-backed evidence before writing relationships
- attach confidence scores to all writes
- run a verification pass after expansion: ask the LLM to check for contradictions with existing graph data
- merge aliases instead of creating near-duplicate nodes

### Evaluation

#### Benchmark question set

Create a set of 20–30 questions with expected Cypher and expected results. Run after prompt or schema changes to catch regressions.

Track failure modes:

- invalid Cypher generation
- wrong entity resolution
- empty results when data exists
- hallucinated relationships
- low-quality answer synthesis

## Recommended order

1. **Few-shot examples in prompts** — immediate Cypher accuracy improvement
2. **Error recovery coaching** — fewer failed queries
3. **Relationship properties** (`since`, `notes`) — richer answers
4. **Provenance on graph expansion writes** — trustworthy knowledge graph
5. **Node `description` property** — foundation for vector search
6. **Vector index + hybrid retrieval** — handles semantic/fuzzy queries
7. **More node types** (`Product`, `Technology`) — 10x richer graph
8. **Benchmark question set** — makes all future changes safer
