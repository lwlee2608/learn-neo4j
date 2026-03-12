# Neo4j Concepts

## Core Building Blocks

### Nodes

Entities in the graph (similar to rows in a relational DB). Each node can have one or more **labels** that describe its type.

```
(:Company)
(:Company:ChipDesigner)   -- multiple labels
```

### Relationships

Named, directed edges between nodes. They always have a type and a direction.

```
(:Company)-[:SUPPLIES_EQUIPMENT_TO]->(:Company)
(:Company)-[:MANUFACTURES_FOR]->(:Company)
(:Company)-[:SUPPLIES_CHIPS_TO]->(:Company)
(:Company)-[:PROVIDES_CLOUD_FOR]->(:Company)
(:Company)-[:COMPETES_WITH]->(:Company)
```

### Properties

Key-value pairs stored on both nodes and relationships.

```
(:Company {name: "NVIDIA", type: "chip_designer", founded: 1993, hq: "Santa Clara"})
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
- Supply chain analysis
- Network/IT infrastructure mapping
- Dependency analysis
