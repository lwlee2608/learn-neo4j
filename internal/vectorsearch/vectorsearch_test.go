package vectorsearch

import (
	"context"
	"os"
	"testing"

	llm "github.com/lwlee2608/learn-neo4j/pkg/ai"
)

func TestEmbedding(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	client := llm.NewOpenAIService(apiKey, "https://api.openai.com/v1")
	vs := New(nil, client, "text-embedding-3-small")

	ctx := context.Background()
	vec, err := vs.embedding.Embedding(ctx, "Taiwan semiconductor manufacturer", vs.model)
	if err != nil {
		t.Fatalf("embedding call failed: %v", err)
	}

	if len(vec) != 1536 {
		t.Fatalf("expected 1536 dimensions, got %d", len(vec))
	}

	// Verify values are non-zero (a valid embedding should have non-zero components)
	nonZero := 0
	for _, v := range vec {
		if v != 0 {
			nonZero++
		}
	}
	if nonZero == 0 {
		t.Fatal("embedding is all zeros")
	}
}
