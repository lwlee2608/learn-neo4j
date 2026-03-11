# Neo4j Concepts

## Core Building Blocks

### Nodes

Entities in the graph (similar to rows in a relational DB). Each node can have one or more **labels** that describe its type.

```
(:Person)
(:Movie)
(:Person:Actor)   -- multiple labels
```

### Relationships

Named, directed edges between nodes. They always have a type and a direction.

```
(:Person)-[:ACTED_IN]->(:Movie)
(:Person)-[:DIRECTED]->(:Movie)
(:Person)-[:FRIENDS_WITH]->(:Person)
```

### Properties

Key-value pairs stored on both nodes and relationships.

```
(:Person {name: "Alice", born: 1990})
[:ACTED_IN {role: "Neo"}]
```

## Graph vs Relational

| Relational | Graph |
|-----------|-------|
| Table | Label |
| Row | Node |
| Column | Property |
| Foreign Key / Join Table | Relationship |
| JOIN query | Graph traversal |

The key advantage: relationships are first-class citizens, not join tables. Traversing connections is O(1) per hop, regardless of dataset size.

## When to Use Neo4j

- Social networks (friends-of-friends)
- Recommendation engines
- Fraud detection (finding suspicious patterns)
- Knowledge graphs
- Network/IT infrastructure mapping
- Dependency analysis
