package vectorsearch

import (
	"context"
	"fmt"

	"github.com/lwlee2608/learn-neo4j/pkg/ai"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type SearchResult struct {
	Name  string
	Score float64
}

type VectorSearch struct {
	driver    neo4j.DriverWithContext
	embedding ai.Embedding
	model     string
}

func New(driver neo4j.DriverWithContext, embedding ai.Embedding, model string) *VectorSearch {
	return &VectorSearch{
		driver:    driver,
		embedding: embedding,
		model:     model,
	}
}

// EmbedAndStore generates an embedding for text and stores it on the named node.
func (v *VectorSearch) EmbedAndStore(ctx context.Context, label, name, text string) error {
	vec, err := v.embedding.Embedding(ctx, text, v.model)
	if err != nil {
		return fmt.Errorf("embed %q: %w", name, err)
	}

	// Convert []float32 to []float64 for Neo4j driver compatibility.
	f64 := make([]float64, len(vec))
	for i, val := range vec {
		f64[i] = float64(val)
	}

	cypher := fmt.Sprintf("MATCH (n:%s {name: $name}) SET n.embedding = $embedding", label)
	_, err = neo4j.ExecuteQuery(ctx, v.driver, cypher,
		map[string]any{"name": name, "embedding": f64},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	return err
}

// Search embeds the query and returns the top-k similar nodes.
func (v *VectorSearch) Search(ctx context.Context, query string, topK int) ([]SearchResult, error) {
	vec, err := v.embedding.Embedding(ctx, query, v.model)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	f64 := make([]float64, len(vec))
	for i, val := range vec {
		f64[i] = float64(val)
	}

	result, err := neo4j.ExecuteQuery(ctx, v.driver,
		`CALL db.index.vector.queryNodes('company_embedding', $topK, $queryVector)
		 YIELD node, score
		 RETURN node.name AS name, score
		 ORDER BY score DESC`,
		map[string]any{"topK": topK, "queryVector": f64},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, record := range result.Records {
		name, _ := record.Get("name")
		score, _ := record.Get("score")
		results = append(results, SearchResult{
			Name:  name.(string),
			Score: score.(float64),
		})
	}
	return results, nil
}

// CreateIndex creates the vector index for company embeddings.
func (v *VectorSearch) CreateIndex(ctx context.Context, dimensions int) error {
	cypher := fmt.Sprintf(
		`CREATE VECTOR INDEX company_embedding IF NOT EXISTS
		 FOR (c:Company) ON (c.embedding)
		 OPTIONS {indexConfig: {
		   `+"`vector.dimensions`"+`: %d,
		   `+"`vector.similarity_function`"+`: 'cosine'
		 }}`, dimensions)

	_, err := neo4j.ExecuteQuery(ctx, v.driver, cypher, nil,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	return err
}
