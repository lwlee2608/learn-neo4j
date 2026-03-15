package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/lwlee2608/learn-neo4j/internal/vectorsearch"
	llm "github.com/lwlee2608/learn-neo4j/pkg/ai"
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
		{"name": "NVIDIA", "type": "chip_designer", "founded": 1993, "hq": "Santa Clara",
			"description": "American chip designer headquartered in Santa Clara. Designs GPUs and AI accelerators (H100, A100) used for training and inference of large language models. Dominates the AI chip market. Chips are manufactured by TSMC in Taiwan."},
		{"name": "AMD", "type": "chip_designer", "founded": 1969, "hq": "Santa Clara",
			"description": "American semiconductor company designing CPUs and GPUs. Competes with NVIDIA in AI accelerators (MI300X) and with Intel in server CPUs. Chips are fabricated by TSMC in Taiwan."},
		{"name": "Intel", "type": "chip_designer", "founded": 1968, "hq": "Santa Clara",
			"description": "American semiconductor company designing CPUs and AI accelerators (Gaudi). Has its own fabs but also uses TSMC for advanced nodes. Competes with NVIDIA and AMD in the AI chip space."},
		{"name": "TSMC", "type": "manufacturer", "founded": 1987, "hq": "Hsinchu",
			"description": "Taiwan-based semiconductor manufacturer. World's largest chip foundry, fabricating chips for NVIDIA, AMD, and Intel. Located in Hsinchu, Taiwan. Critical single point of failure in the global AI supply chain due to geographic concentration."},
		{"name": "Samsung Foundry", "type": "manufacturer", "founded": 1969, "hq": "Suwon",
			"description": "South Korean semiconductor foundry based in Suwon. Manufactures chips for Google and AWS. Competes with TSMC but has smaller market share in advanced nodes."},
		{"name": "ASML", "type": "equipment_supplier", "founded": 1984, "hq": "Veldhoven",
			"description": "Dutch company building extreme ultraviolet (EUV) lithography machines used by chip foundries like TSMC and Samsung. Sole supplier of EUV equipment, making it a critical bottleneck in semiconductor manufacturing."},
		{"name": "OpenAI", "type": "ai_lab", "founded": 2015, "hq": "San Francisco",
			"description": "San Francisco-based AI research lab. Creator of GPT-4 and ChatGPT. Heavily reliant on NVIDIA GPUs for training. Uses Microsoft Azure, CoreWeave, and Oracle Cloud for compute infrastructure."},
		{"name": "Anthropic", "type": "ai_lab", "founded": 2021, "hq": "San Francisco",
			"description": "San Francisco-based AI safety company. Creator of Claude. Uses NVIDIA GPUs and AWS cloud infrastructure for training and serving models. Competes with OpenAI and Google DeepMind."},
		{"name": "Google DeepMind", "type": "ai_lab", "founded": 2010, "hq": "London",
			"description": "London-based AI research lab owned by Google. Creator of Gemini models. Uses Google Cloud infrastructure and custom TPU chips alongside Intel processors. Competes with OpenAI and Anthropic."},
		{"name": "Meta AI", "type": "ai_lab", "founded": 2013, "hq": "Menlo Park",
			"description": "Menlo Park-based AI research division of Meta. Develops Llama open-source models. Uses NVIDIA and AMD GPUs with CoreWeave cloud infrastructure. Competes with OpenAI, Anthropic, and Google DeepMind."},
		{"name": "Moonshot AI", "type": "ai_lab", "founded": 2023, "hq": "Beijing",
			"description": "Beijing-based Chinese AI startup. Develops large language models for the Chinese market. Uses NVIDIA GPUs. Competes with z.ai in the Chinese AI space."},
		{"name": "z.ai", "type": "ai_lab", "founded": 2023, "hq": "San Francisco",
			"description": "San Francisco-based AI startup founded in 2023. Develops AI models and competes with Moonshot AI. Uses NVIDIA GPUs for compute."},
		{"name": "AWS", "type": "cloud_provider", "founded": 2006, "hq": "Seattle",
			"description": "Amazon's cloud computing platform based in Seattle. Provides cloud infrastructure for Anthropic. Designs custom Trainium and Inferentia AI chips manufactured by Samsung. Competes with Azure, Google Cloud, Oracle Cloud, and CoreWeave."},
		{"name": "Microsoft Azure", "type": "cloud_provider", "founded": 2010, "hq": "Redmond",
			"description": "Microsoft's cloud computing platform based in Redmond. Major cloud provider for OpenAI. Competes with AWS, Google Cloud, Oracle Cloud, and CoreWeave in cloud infrastructure."},
		{"name": "Google Cloud", "type": "cloud_provider", "founded": 2008, "hq": "Sunnyvale",
			"description": "Google's cloud computing platform based in Sunnyvale. Provides infrastructure for Google DeepMind. Offers TPU chips for AI workloads. Competes with AWS, Azure, Oracle Cloud, and CoreWeave."},
		{"name": "Oracle Cloud", "type": "cloud_provider", "founded": 2016, "hq": "Austin",
			"description": "Oracle's cloud infrastructure platform based in Austin. Provides cloud compute for OpenAI training runs. Competes with AWS, Azure, Google Cloud, and CoreWeave."},
		{"name": "CoreWeave", "type": "cloud_provider", "founded": 2017, "hq": "Roseland",
			"description": "GPU-focused cloud provider based in Roseland, New Jersey. Specializes in NVIDIA GPU clusters for AI workloads. Provides compute for OpenAI and Meta AI. Competes with AWS, Azure, Google Cloud, and Oracle Cloud."},
	}
	for _, c := range companies {
		run(ctx, driver, "CREATE (:Company {name: $name, type: $type, founded: $founded, hq: $hq, description: $description})", c)
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
		{"provider": "CoreWeave", "client": "OpenAI"},
		{"provider": "Microsoft Azure", "client": "OpenAI"},
		{"provider": "Google Cloud", "client": "Google DeepMind"},
		{"provider": "Oracle Cloud", "client": "OpenAI"},
		{"provider": "CoreWeave", "client": "Meta AI"},
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
		{"company": "AWS", "competitor": "CoreWeave"},
		{"company": "Microsoft Azure", "competitor": "Google Cloud"},
		{"company": "Microsoft Azure", "competitor": "Oracle Cloud"},
		{"company": "Microsoft Azure", "competitor": "CoreWeave"},
		{"company": "Google Cloud", "competitor": "Oracle Cloud"},
		{"company": "Google Cloud", "competitor": "CoreWeave"},
		{"company": "Oracle Cloud", "competitor": "CoreWeave"},
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

	// Generate embeddings if OpenRouter API key is available
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENROUTER_APIKEY")
	}
	if apiKey != "" {
		baseURL := envOrDefault("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1")
		embModel := envOrDefault("EMBEDDING_MODEL", "text-embedding-3-small")
		dimensions := 1536

		aiClient := llm.NewOpenAIService(apiKey, baseURL)
		vs := vectorsearch.New(driver, aiClient, embModel)

		slog.Info("Creating vector index", "dimensions", dimensions)
		if err := vs.CreateIndex(ctx, dimensions); err != nil {
			slog.Error("Failed to create vector index", "error", err)
			os.Exit(1)
		}

		slog.Info("Generating embeddings for companies")
		for _, c := range companies {
			name := c["name"].(string)
			desc := c["description"].(string)
			if err := vs.EmbedAndStore(ctx, "Company", name, desc); err != nil {
				slog.Error("Failed to embed company", "name", name, "error", err)
				os.Exit(1)
			}
			slog.Info("Embedded company", "name", name)
		}
		slog.Info("Embeddings complete")
	} else {
		slog.Warn("OPENROUTER_API_KEY not set, skipping embedding generation")
	}

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
