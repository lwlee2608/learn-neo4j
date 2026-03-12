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
	run(ctx, driver, "CREATE INDEX IF NOT EXISTS FOR (ch:Chip) ON (ch.name)", nil)
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

	// Chips
	chips := []map[string]any{
		{"name": "H100", "architecture": "Hopper", "year": 2022, "transistor_nm": 4},
		{"name": "A100", "architecture": "Ampere", "year": 2020, "transistor_nm": 7},
		{"name": "H200", "architecture": "Hopper", "year": 2024, "transistor_nm": 4},
		{"name": "B200", "architecture": "Blackwell", "year": 2024, "transistor_nm": 4},
		{"name": "MI300X", "architecture": "CDNA 3", "year": 2023, "transistor_nm": 5},
		{"name": "Gaudi 3", "architecture": "Gaudi", "year": 2024, "transistor_nm": 5},
		{"name": "TPU v5e", "architecture": "TPU", "year": 2023, "transistor_nm": 7},
		{"name": "Trainium2", "architecture": "Trainium", "year": 2024, "transistor_nm": 3},
	}
	for _, ch := range chips {
		run(ctx, driver, "CREATE (:Chip {name: $name, architecture: $architecture, year: $year, transistor_nm: $transistor_nm})", ch)
	}
	slog.Info("Created chips", "count", len(chips))

	// DESIGNED relationships
	designed := []map[string]any{
		{"company": "NVIDIA", "chip": "H100"},
		{"company": "NVIDIA", "chip": "A100"},
		{"company": "NVIDIA", "chip": "H200"},
		{"company": "NVIDIA", "chip": "B200"},
		{"company": "AMD", "chip": "MI300X"},
		{"company": "Intel", "chip": "Gaudi 3"},
		{"company": "Google DeepMind", "chip": "TPU v5e"},
		{"company": "AWS", "chip": "Trainium2"},
	}
	for _, d := range designed {
		run(ctx, driver,
			`MATCH (c:Company {name: $company})
			 MATCH (ch:Chip {name: $chip})
			 CREATE (c)-[:DESIGNED]->(ch)`, d)
	}
	slog.Info("Created DESIGNED relationships", "count", len(designed))

	// MANUFACTURES relationships
	manufactures := []map[string]any{
		{"company": "TSMC", "chip": "H100"},
		{"company": "TSMC", "chip": "A100"},
		{"company": "TSMC", "chip": "H200"},
		{"company": "TSMC", "chip": "B200"},
		{"company": "TSMC", "chip": "MI300X"},
		{"company": "TSMC", "chip": "Gaudi 3"},
		{"company": "Samsung Foundry", "chip": "TPU v5e"},
		{"company": "Samsung Foundry", "chip": "Trainium2"},
	}
	for _, m := range manufactures {
		run(ctx, driver,
			`MATCH (c:Company {name: $company})
			 MATCH (ch:Chip {name: $chip})
			 CREATE (c)-[:MANUFACTURES]->(ch)`, m)
	}
	slog.Info("Created MANUFACTURES relationships", "count", len(manufactures))

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

	// USES relationships
	uses := []map[string]any{
		{"company": "OpenAI", "chip": "H100"},
		{"company": "OpenAI", "chip": "A100"},
		{"company": "Anthropic", "chip": "H100"},
		{"company": "Anthropic", "chip": "H200"},
		{"company": "Google DeepMind", "chip": "TPU v5e"},
		{"company": "Meta AI", "chip": "H100"},
		{"company": "Meta AI", "chip": "A100"},
		{"company": "Moonshot AI", "chip": "H100"},
		{"company": "z.ai", "chip": "H100"},
	}
	for _, u := range uses {
		run(ctx, driver,
			`MATCH (c:Company {name: $company})
			 MATCH (ch:Chip {name: $chip})
			 CREATE (c)-[:USES]->(ch)`, u)
	}
	slog.Info("Created USES relationships", "count", len(uses))

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
