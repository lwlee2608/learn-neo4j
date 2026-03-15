package vectorsearch

import (
	"context"
	"os"
	"testing"

	llm "github.com/lwlee2608/learn-neo4j/pkg/ai"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestVectorSearch(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	neo4jURI := envOrDefault("NEO4J_URI", "bolt://localhost:7687")
	neo4jUser := envOrDefault("NEO4J_USERNAME", "neo4j")
	neo4jPass := envOrDefault("NEO4J_PASSWORD", "password")

	driver, err := neo4j.NewDriverWithContext(neo4jURI, neo4j.BasicAuth(neo4jUser, neo4jPass, ""))
	if err != nil {
		t.Fatalf("create driver: %v", err)
	}
	ctx := context.Background()
	defer driver.Close(ctx)

	if err := driver.VerifyConnectivity(ctx); err != nil {
		t.Skip("Neo4j not available: " + err.Error())
	}

	// Clean up test data before and after
	cleanup := func() {
		neo4j.ExecuteQuery(ctx, driver, "MATCH (c:Company) WHERE c.name STARTS WITH '__test_' DETACH DELETE c", nil,
			neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase("neo4j"))
		neo4j.ExecuteQuery(ctx, driver, "DROP INDEX test_company_embedding IF EXISTS", nil,
			neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase("neo4j"))
	}
	cleanup()
	t.Cleanup(cleanup)

	// Seed test companies — mix of relevant (semiconductor) and irrelevant (coffee, sportswear)
	companies := []struct {
		name string
		desc string
	}{
		{"__test_TSMC", "Taiwan-based semiconductor manufacturer, world's largest chip foundry located in Hsinchu, Taiwan"},
		{"__test_NVIDIA", "American chip designer making GPUs and AI accelerators for training large language models"},
		{"__test_AWS", "Amazon cloud computing platform providing infrastructure for AI companies"},
		{"__test_Starbucks", "American coffeehouse chain selling coffee drinks and food items worldwide"},
		{"__test_Nike", "American sportswear company designing athletic shoes and apparel"},
	}

	for _, c := range companies {
		_, err := neo4j.ExecuteQuery(ctx, driver,
			"CREATE (:Company {name: $name, description: $desc})",
			map[string]any{"name": c.name, "desc": c.desc},
			neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase("neo4j"))
		if err != nil {
			t.Fatalf("create %s: %v", c.name, err)
		}
	}

	client := llm.NewOpenAIService(apiKey, "https://api.openai.com/v1")
	vs := New(driver, client, "text-embedding-3-small")

	// Create vector index
	_, err = neo4j.ExecuteQuery(ctx, driver,
		`CREATE VECTOR INDEX test_company_embedding IF NOT EXISTS
		 FOR (c:Company) ON (c.embedding)
		 OPTIONS {indexConfig: {`+
			"`vector.dimensions`"+`: 1536,`+
			"`vector.similarity_function`"+`: 'cosine'
		 }}`, nil,
		neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase("neo4j"))
	if err != nil {
		t.Fatalf("create index: %v", err)
	}

	// Generate embeddings
	for _, c := range companies {
		if err := vs.EmbedAndStore(ctx, "Company", c.name, c.desc); err != nil {
			t.Fatalf("embed %s: %v", c.name, err)
		}
	}

	// Search for semiconductor-related query
	results, err := vs.searchWithIndex(ctx, "test_company_embedding", "semiconductor chip manufacturing in Taiwan", 5)
	if err != nil {
		t.Fatalf("search: %v", err)
	}

	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}

	// Build score map for assertions
	scores := map[string]float64{}
	for _, r := range results {
		scores[r.Name] = r.Score
		t.Logf("  %s (score: %.4f)", r.Name, r.Score)
	}

	// TSMC should be the top result (most relevant to "semiconductor chip manufacturing in Taiwan")
	if results[0].Name != "__test_TSMC" {
		t.Errorf("expected __test_TSMC as top result, got %s", results[0].Name)
	}

	// Semiconductor companies should score higher than unrelated companies
	if scores["__test_TSMC"] <= scores["__test_Starbucks"] {
		t.Error("expected TSMC to score higher than Starbucks")
	}
	if scores["__test_NVIDIA"] <= scores["__test_Nike"] {
		t.Error("expected NVIDIA to score higher than Nike")
	}
	if scores["__test_TSMC"] <= scores["__test_Nike"] {
		t.Error("expected TSMC to score higher than Nike")
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
