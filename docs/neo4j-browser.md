# Neo4j Browser (localhost:7474)

## Connecting

- **URL**: `bolt://localhost:7687` (pre-filled)
- **Username**: `neo4j`
- **Password**: `password`

## The Interface

The **command bar** at the top is where you type Cypher queries. Results appear below as interactive visualizations.

## Example Queries

Run `make seed` first to populate the database.

### See everything in the graph

```cypher
MATCH (n)-[r]->(m) RETURN n, r, m
```

Shows all nodes and relationships as an interactive graph. Drag nodes around, click them to see properties.

### Explore a specific company

```cypher
MATCH (c:Company {name: "NVIDIA"})-[r]-(other) RETURN c, r, other
```

### Full supply chain for an AI lab

```cypher
MATCH path = (supplier)-[*1..3]->(c:Company {name: "Anthropic"})
RETURN path
```

Traces the chain: ASML → TSMC → NVIDIA → Anthropic, plus AWS → Anthropic.

### All competitors in the graph

```cypher
MATCH (a:Company)-[:COMPETES_WITH]->(b:Company)
RETURN a, b
```

### Companies by type

```cypher
MATCH (c:Company)
RETURN c.type AS type, collect(c.name) AS companies
ORDER BY type
```

### Shortest path between two companies

```cypher
MATCH path = shortestPath(
  (a:Company {name: "ASML"})-[*..10]-(b:Company {name: "OpenAI"})
)
RETURN path
```

Shows the supply chain path: ASML → TSMC → NVIDIA → OpenAI.

### Who has the most relationships?

```cypher
MATCH (c:Company)-[r]-()
RETURN c.name, count(r) AS connections
ORDER BY connections DESC
```

This shows as a table — click the table icon on the result panel to switch views.

## Result Views

Each query result has icons to switch between:

- **Graph** — interactive node/relationship visualization
- **Table** — tabular data
- **Text** — plain text output
- **Code** — raw response

## Interacting with the Graph View

- **Click a node** to see its properties
- **Double-click a node** to expand its relationships
- **Drag nodes** to rearrange the layout

## Built-in Guides

Type these in the command bar:

- `:help` — general help
- `:help cypher` — Cypher cheat sheet

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| Ctrl+Enter | Run query |
| Ctrl+Up/Down | Navigate query history |
| Escape | Clear the command bar |

## Exporting Results

From the result panel you can export as CSV, JSON, PNG, or SVG.
