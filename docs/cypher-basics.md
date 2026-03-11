# Cypher Query Language

Cypher is Neo4j's declarative query language (like SQL for graphs). It uses ASCII-art patterns to describe graph shapes.

## CREATE — Insert Data

```cypher
-- Create a node
CREATE (m:Movie {title: "The Matrix", released: 1999, tagline: "Welcome to the Real World"})

-- Create a person
CREATE (p:Person {name: "Keanu Reeves", born: 1964})

-- Create a relationship (nodes must exist first)
MATCH (p:Person {name: "Keanu Reeves"})
MATCH (m:Movie {title: "The Matrix"})
CREATE (p)-[:ACTED_IN {role: "Neo"}]->(m)
```

## MATCH — Read Data

```cypher
-- Find all movies
MATCH (m:Movie) RETURN m

-- Find a specific person
MATCH (p:Person {name: "Keanu Reeves"}) RETURN p

-- Find who acted in a movie
MATCH (p:Person)-[r:ACTED_IN]->(m:Movie {title: "The Matrix"})
RETURN p.name, r.role

-- Find all movies a person acted in
MATCH (p:Person {name: "Keanu Reeves"})-[:ACTED_IN]->(m:Movie)
RETURN m.title, m.released

-- Find co-actors (2 hops)
MATCH (p:Person {name: "Keanu Reeves"})-[:ACTED_IN]->(m)<-[:ACTED_IN]-(coactor)
RETURN DISTINCT coactor.name
```

## SET — Update Data

```cypher
-- Update a property
MATCH (p:Person {name: "Keanu Reeves"})
SET p.born = 1964

-- Add a new property
MATCH (m:Movie {title: "The Matrix"})
SET m.budget = 63000000
```

## DELETE — Remove Data

```cypher
-- Delete a node (must have no relationships)
MATCH (p:Person {name: "Keanu Reeves"})
DELETE p

-- Delete a node and all its relationships
MATCH (p:Person {name: "Keanu Reeves"})
DETACH DELETE p

-- Delete a specific relationship
MATCH (p:Person {name: "Keanu Reeves"})-[r:ACTED_IN]->(m:Movie {title: "The Matrix"})
DELETE r
```

## MERGE — Create If Not Exists

```cypher
-- Like an upsert: creates the node only if it doesn't exist
MERGE (p:Person {name: "Keanu Reeves"})
ON CREATE SET p.born = 1964
ON MATCH SET p.lastSeen = timestamp()
```

## WHERE — Filtering

```cypher
-- Filter with conditions
MATCH (m:Movie)
WHERE m.released > 2000
RETURN m.title, m.released
ORDER BY m.released

-- Pattern filtering
MATCH (p:Person)
WHERE NOT (p)-[:ACTED_IN]->()
RETURN p.name
```

## Aggregation

```cypher
-- Count movies per person
MATCH (p:Person)-[:ACTED_IN]->(m:Movie)
RETURN p.name, count(m) AS movieCount
ORDER BY movieCount DESC

-- Collect into a list
MATCH (p:Person)-[r:ACTED_IN]->(m:Movie {title: "The Matrix"})
RETURN m.title, collect({name: p.name, role: r.role}) AS cast
```

## Tips

- Use the Neo4j Browser (http://localhost:7474) to experiment with queries interactively
- `EXPLAIN` before a query shows the execution plan without running it
- `PROFILE` before a query runs it and shows actual performance metrics
- Create indexes for frequently queried properties:
  ```cypher
  CREATE INDEX FOR (m:Movie) ON (m.title)
  CREATE INDEX FOR (p:Person) ON (p.name)
  ```
