# Cypher Query Language

Cypher is Neo4j's declarative query language (like SQL for graphs). It uses ASCII-art patterns to describe graph shapes.

## CREATE — Insert Data

```cypher
-- Create a node
CREATE (c:Company {name: "NVIDIA", type: "chip_designer", founded: 1993, hq: "Santa Clara"})

-- Create a relationship (nodes must exist first)
MATCH (s:Company {name: "NVIDIA"})
MATCH (c:Company {name: "OpenAI"})
CREATE (s)-[:SUPPLIES_CHIPS_TO]->(c)
```

## MATCH — Read Data

```cypher
-- Find all companies
MATCH (c:Company) RETURN c

-- Find a specific company
MATCH (c:Company {name: "NVIDIA"}) RETURN c

-- Find who TSMC manufactures for
MATCH (c:Company {name: "TSMC"})-[:MANUFACTURES_FOR]->(client:Company)
RETURN client.name

-- Find the full supply chain for an AI lab
MATCH (c:Company {name: "OpenAI"})<-[r]-(supplier:Company)
RETURN supplier.name, type(r) AS relationship

-- Find competitors
MATCH (c:Company {name: "NVIDIA"})-[:COMPETES_WITH]-(competitor:Company)
RETURN competitor.name
```

## SET — Update Data

```cypher
-- Update a property
MATCH (c:Company {name: "NVIDIA"})
SET c.founded = 1993

-- Add a new property
MATCH (c:Company {name: "NVIDIA"})
SET c.market_cap = 3000000000000
```

## DELETE — Remove Data

```cypher
-- Delete a node (must have no relationships)
MATCH (c:Company {name: "NVIDIA"})
DELETE c

-- Delete a node and all its relationships
MATCH (c:Company {name: "NVIDIA"})
DETACH DELETE c

-- Delete a specific relationship
MATCH (c:Company {name: "NVIDIA"})-[r:SUPPLIES_CHIPS_TO]->(o:Company {name: "OpenAI"})
DELETE r
```

## MERGE — Create If Not Exists

```cypher
-- Like an upsert: creates the node only if it doesn't exist
MERGE (c:Company {name: "NVIDIA"})
ON CREATE SET c.type = "chip_designer", c.founded = 1993
ON MATCH SET c.lastSeen = timestamp()
```

## WHERE — Filtering

```cypher
-- Filter with conditions
MATCH (c:Company)
WHERE c.founded > 2000
RETURN c.name, c.type, c.founded
ORDER BY c.founded

-- Pattern filtering: companies with no outgoing relationships
MATCH (c:Company)
WHERE NOT (c)-[]->()
RETURN c.name
```

## Aggregation

```cypher
-- Count relationships per company
MATCH (c:Company)-[r]->()
RETURN c.name, type(r) AS relationship, count(*) AS count
ORDER BY count DESC

-- Collect into a list
MATCH (c:Company {name: "NVIDIA"})-[r]->(other:Company)
RETURN type(r) AS relationship, collect(other.name) AS companies
```

## Tips

- Use the Neo4j Browser (http://localhost:7474) to experiment with queries interactively
- `EXPLAIN` before a query shows the execution plan without running it
- `PROFILE` before a query runs it and shows actual performance metrics
- Create indexes for frequently queried properties:
  ```cypher
  CREATE INDEX FOR (c:Company) ON (c.name)
  ```
