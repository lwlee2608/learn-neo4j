package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	uri := envOrDefault("NEO4J_URI", "bolt://localhost:7687")
	username := envOrDefault("NEO4J_USERNAME", "neo4j")
	password := envOrDefault("NEO4J_PASSWORD", "password")

	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		slog.Error("Failed to create driver", "error", err)
		os.Exit(1)
	}
	defer driver.Close(context.Background())

	ctx := context.Background()

	if err := driver.VerifyConnectivity(ctx); err != nil {
		slog.Error("Failed to connect to Neo4j", "error", err)
		os.Exit(1)
	}
	slog.Info("Connected to Neo4j", "uri", uri)

	// Clear existing data
	run(ctx, driver, "MATCH (n) DETACH DELETE n", nil)
	slog.Info("Cleared existing data")

	// Create indexes
	run(ctx, driver, "CREATE INDEX IF NOT EXISTS FOR (c:Company) ON (c.name)", nil)
	slog.Info("Created indexes")

	// Companies
	companies := []map[string]any{
		{"name": "NVIDIA", "type": "chip_designer", "founded": 1993, "hq": "Santa Clara"},
		{"name": "AMD", "type": "chip_designer", "founded": 1969, "hq": "Santa Clara"},
		{"name": "Intel", "type": "chip_designer", "founded": 1968, "hq": "Santa Clara"},
		{"name": "TSMC", "type": "manufacturer", "founded": 1987, "hq": "Hsinchu"},
		{"name": "Samsung Foundry", "type": "manufacturer", "founded": 1969, "hq": "Suwon"},
		{"name": "ASML", "type": "equipment_supplier", "founded": 1984, "hq": "Veldhoven"},
		{"name": "OpenAI", "type": "ai_lab", "founded": 2015, "hq": "San Francisco"},
		{"name": "Anthropic", "type": "ai_lab", "founded": 2021, "hq": "San Francisco"},
		{"name": "Google DeepMind", "type": "ai_lab", "founded": 2010, "hq": "London"},
		{"name": "Meta AI", "type": "ai_lab", "founded": 2013, "hq": "Menlo Park"},
		{"name": "Moonshot AI", "type": "ai_lab", "founded": 2023, "hq": "Beijing"},
		{"name": "z.ai", "type": "ai_lab", "founded": 2023, "hq": "San Francisco"},
		{"name": "AWS", "type": "cloud_provider", "founded": 2006, "hq": "Seattle"},
		{"name": "Microsoft Azure", "type": "cloud_provider", "founded": 2010, "hq": "Redmond"},
		{"name": "Google Cloud", "type": "cloud_provider", "founded": 2008, "hq": "Sunnyvale"},
		{"name": "Oracle Cloud", "type": "cloud_provider", "founded": 2016, "hq": "Austin"},
	}
	for _, c := range companies {
		run(ctx, driver, "CREATE (:Company {name: $name, type: $type, founded: $founded, hq: $hq})", c)
	}
	slog.Info("Created companies", "count", len(companies))

	// SUPPLIES_EQUIPMENT_TO relationships
	suppliesEquipment := []map[string]any{
		{"supplier": "ASML", "recipient": "TSMC"},
		{"supplier": "ASML", "recipient": "Samsung Foundry"},
	}
	for _, s := range suppliesEquipment {
		run(ctx, driver,
			`MATCH (s:Company {name: $supplier})
			 MATCH (r:Company {name: $recipient})
			 CREATE (s)-[:SUPPLIES_EQUIPMENT_TO]->(r)`, s)
	}
	slog.Info("Created SUPPLIES_EQUIPMENT_TO relationships", "count", len(suppliesEquipment))

	// MANUFACTURES_FOR relationships
	manufacturesFor := []map[string]any{
		{"manufacturer": "TSMC", "client": "NVIDIA"},
		{"manufacturer": "TSMC", "client": "AMD"},
		{"manufacturer": "TSMC", "client": "Intel"},
		{"manufacturer": "Samsung Foundry", "client": "Google DeepMind"},
		{"manufacturer": "Samsung Foundry", "client": "AWS"},
	}
	for _, m := range manufacturesFor {
		run(ctx, driver,
			`MATCH (m:Company {name: $manufacturer})
			 MATCH (c:Company {name: $client})
			 CREATE (m)-[:MANUFACTURES_FOR]->(c)`, m)
	}
	slog.Info("Created MANUFACTURES_FOR relationships", "count", len(manufacturesFor))

	// SUPPLIES_CHIPS_TO relationships
	suppliesChips := []map[string]any{
		{"supplier": "NVIDIA", "client": "OpenAI"},
		{"supplier": "NVIDIA", "client": "Anthropic"},
		{"supplier": "NVIDIA", "client": "Meta AI"},
		{"supplier": "NVIDIA", "client": "Moonshot AI"},
		{"supplier": "NVIDIA", "client": "z.ai"},
		{"supplier": "AMD", "client": "Meta AI"},
		{"supplier": "Intel", "client": "Google DeepMind"},
	}
	for _, s := range suppliesChips {
		run(ctx, driver,
			`MATCH (s:Company {name: $supplier})
			 MATCH (c:Company {name: $client})
			 CREATE (s)-[:SUPPLIES_CHIPS_TO]->(c)`, s)
	}
	slog.Info("Created SUPPLIES_CHIPS_TO relationships", "count", len(suppliesChips))

	// PROVIDES_CLOUD_FOR relationships
	providesCloud := []map[string]any{
		{"provider": "AWS", "client": "Anthropic"},
		{"provider": "Microsoft Azure", "client": "OpenAI"},
		{"provider": "Google Cloud", "client": "Google DeepMind"},
		{"provider": "Oracle Cloud", "client": "OpenAI"},
	}
	for _, p := range providesCloud {
		run(ctx, driver,
			`MATCH (p:Company {name: $provider})
			 MATCH (c:Company {name: $client})
			 CREATE (p)-[:PROVIDES_CLOUD_FOR]->(c)`, p)
	}
	slog.Info("Created PROVIDES_CLOUD_FOR relationships", "count", len(providesCloud))

	// COMPETES_WITH relationships
	competesWith := []map[string]any{
		{"company": "NVIDIA", "competitor": "AMD"},
		{"company": "NVIDIA", "competitor": "Intel"},
		{"company": "AMD", "competitor": "Intel"},
		{"company": "TSMC", "competitor": "Samsung Foundry"},
		{"company": "AWS", "competitor": "Microsoft Azure"},
		{"company": "AWS", "competitor": "Google Cloud"},
		{"company": "AWS", "competitor": "Oracle Cloud"},
		{"company": "Microsoft Azure", "competitor": "Google Cloud"},
		{"company": "Microsoft Azure", "competitor": "Oracle Cloud"},
		{"company": "Google Cloud", "competitor": "Oracle Cloud"},
		{"company": "OpenAI", "competitor": "Anthropic"},
		{"company": "OpenAI", "competitor": "Google DeepMind"},
		{"company": "OpenAI", "competitor": "Meta AI"},
		{"company": "Anthropic", "competitor": "Google DeepMind"},
		{"company": "Anthropic", "competitor": "Meta AI"},
		{"company": "Google DeepMind", "competitor": "Meta AI"},
		{"company": "Moonshot AI", "competitor": "z.ai"},
	}
	for _, c := range competesWith {
		run(ctx, driver,
			`MATCH (a:Company {name: $company})
			 MATCH (b:Company {name: $competitor})
			 CREATE (a)-[:COMPETES_WITH]->(b)`, c)
	}
	slog.Info("Created COMPETES_WITH relationships", "count", len(competesWith))

	slog.Info("Seed complete!")
}

func run(ctx context.Context, driver neo4j.DriverWithContext, cypher string, params map[string]any) {
	_, err := neo4j.ExecuteQuery(ctx, driver, cypher, params,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	if err != nil {
		slog.Error("Query failed", "cypher", cypher, "error", err)
		os.Exit(1)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
