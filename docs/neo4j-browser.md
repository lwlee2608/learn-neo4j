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
MATCH (n) RETURN n
```

Shows all nodes and relationships as an interactive graph. Drag nodes around, click them to see properties.

### Explore a specific actor

```cypher
MATCH (p:Person {name: "Keanu Reeves"})-[r]->(m) RETURN p, r, m
```

### Find co-actors

```cypher
MATCH (p:Person {name: "Keanu Reeves"})-[:ACTED_IN]->(m)<-[:ACTED_IN]-(coactor)
RETURN p, m, coactor
```

### All co-actor pairs

```cypher
MATCH (p:Person)-[:ACTED_IN]->(m)<-[:ACTED_IN]-(other)
WHERE p <> other
RETURN p.name, other.name, m.title
ORDER BY p.name
```

### Shortest path between two actors

```cypher
MATCH path = shortestPath(
  (a:Person {name: "Hugo Weaving"})-[*..10]-(b:Person {name: "Keanu Reeves"})
)
RETURN path
```

Note: `shortestPath` only works when a path actually exists. The `[*..10]` sets an upper bound on hops (unbounded `[*]` triggers a warning). With the seed data, some actor clusters are isolated — e.g. Tom Hanks and Keanu Reeves share no connecting path.

### Who has the most relationships?

```cypher
MATCH (p:Person)-[r]->(m:Movie)
RETURN p.name, type(r) AS relationship, count(m) AS count
ORDER BY count DESC
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
- `:play movies` — Neo4j's built-in interactive movie tutorial

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| Ctrl+Enter | Run query |
| Ctrl+Up/Down | Navigate query history |
| Escape | Clear the command bar |

## Exporting Results

From the result panel you can export as CSV, JSON, PNG, or SVG.
