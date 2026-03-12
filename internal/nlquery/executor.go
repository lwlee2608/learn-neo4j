package nlquery

import (
	"context"
	"fmt"

	n "github.com/lwlee2608/learn-neo4j/pkg/neo4j"
	neo4jdb "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func ExecuteReadOnly(ctx context.Context, client *n.Client, plan *Plan) (*QueryResult, error) {
	if err := ValidatePlan(plan, DefaultGraphSchema()); err != nil {
		return nil, err
	}

	result, err := neo4jdb.ExecuteQuery(
		ctx,
		client.Driver,
		plan.Query,
		plan.Params,
		neo4jdb.EagerResultTransformer,
		neo4jdb.ExecuteQueryWithDatabase("neo4j"),
	)
	if err != nil {
		return nil, fmt.Errorf("execute query: %w", err)
	}

	records := make([]map[string]any, 0, len(result.Records))
	for _, record := range result.Records {
		row := make(map[string]any, len(record.Keys))
		for key, value := range record.AsMap() {
			row[key] = normalizeValue(value)
		}
		records = append(records, row)
	}

	return &QueryResult{
		Records: records,
		Count:   len(records),
	}, nil
}

func normalizeValue(value any) any {
	switch v := value.(type) {
	case neo4jdb.Node:
		return map[string]any{
			"element_id": v.ElementId,
			"labels":     v.Labels,
			"props":      normalizeMap(v.Props),
		}
	case neo4jdb.Relationship:
		return map[string]any{
			"element_id":       v.ElementId,
			"start_element_id": v.StartElementId,
			"end_element_id":   v.EndElementId,
			"type":             v.Type,
			"props":            normalizeMap(v.Props),
		}
	case neo4jdb.Path:
		nodes := make([]any, 0, len(v.Nodes))
		for _, node := range v.Nodes {
			nodes = append(nodes, normalizeValue(node))
		}

		relationships := make([]any, 0, len(v.Relationships))
		for _, rel := range v.Relationships {
			relationships = append(relationships, normalizeValue(rel))
		}

		return map[string]any{
			"nodes":         nodes,
			"relationships": relationships,
		}
	case []any:
		items := make([]any, 0, len(v))
		for _, item := range v {
			items = append(items, normalizeValue(item))
		}
		return items
	case map[string]any:
		return normalizeMap(v)
	default:
		return value
	}
}

func normalizeMap(input map[string]any) map[string]any {
	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = normalizeValue(value)
	}
	return output
}
