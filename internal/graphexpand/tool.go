package graphexpand

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/lwlee2608/learn-neo4j/pkg/ai"
	"github.com/lwlee2608/learn-neo4j/pkg/exaai"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
)

type SearchWebTool struct {
	client exaai.Client
}

type searchWebInput struct {
	Query      string `json:"query"`
	NumResults int    `json:"num_results"`
}

func NewSearchWebTool(client exaai.Client) *SearchWebTool {
	return &SearchWebTool{client: client}
}

func (t *SearchWebTool) Definition() openai.ChatCompletionToolUnionParam {
	parameters, err := schemaAsMap(ai.GenerateSchema[searchWebInput]())
	if err != nil {
		panic(fmt.Sprintf("generate search_web schema: %v", err))
	}

	return openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
		Name:        "search_web",
		Description: param.NewOpt("Search the web for recent evidence about companies and relationships"),
		Parameters:  parameters,
	})
}

func (t *SearchWebTool) Execute(ctx context.Context, inputs map[string]any) (ai.ToolResult, error) {
	var input searchWebInput
	if err := decodeInputs(inputs, &input); err != nil {
		return nil, err
	}
	if input.NumResults <= 0 || input.NumResults > 10 {
		input.NumResults = 5
	}

	resp, err := t.client.Search(ctx, exaai.SearchRequest{
		Query:      strings.TrimSpace(input.Query),
		NumResults: input.NumResults,
		Contents: exaai.Contents{
			Text: true,
			Summary: &exaai.Summary{
				Query: input.Query,
			},
		},
		Category: "company",
	})
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	return &ai.SimpleToolResult{ToolContent: string(payload)}, nil
}

type ListCompaniesTool struct {
	store Store
}

func NewListCompaniesTool(store Store) *ListCompaniesTool {
	return &ListCompaniesTool{store: store}
}

func (t *ListCompaniesTool) Definition() openai.ChatCompletionToolUnionParam {
	return openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
		Name:        "list_graph_companies",
		Description: param.NewOpt("List the companies already present in the graph"),
	})
}

func (t *ListCompaniesTool) Execute(ctx context.Context, _ map[string]any) (ai.ToolResult, error) {
	companies, err := t.store.ListCompanies(ctx)
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(map[string]any{"companies": companies})
	if err != nil {
		return nil, err
	}
	return &ai.SimpleToolResult{ToolContent: string(payload)}, nil
}

type ApplyPlanTool struct {
	store Store

	mu         sync.RWMutex
	lastPlan   *Plan
	lastResult *ApplyResult
}

func NewApplyPlanTool(store Store) *ApplyPlanTool {
	return &ApplyPlanTool{store: store}
}

func (t *ApplyPlanTool) Definition() openai.ChatCompletionToolUnionParam {
	parameters, err := schemaAsMap(ai.GenerateSchema[Plan]())
	if err != nil {
		panic(fmt.Sprintf("generate apply_graph_expansion schema: %v", err))
	}

	return openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
		Name:        "apply_graph_expansion",
		Description: param.NewOpt("Create or update companies and relationships in the graph using a validated expansion plan"),
		Parameters:  parameters,
	})
}

func (t *ApplyPlanTool) Execute(ctx context.Context, inputs map[string]any) (ai.ToolResult, error) {
	var plan Plan
	if err := decodeInputs(inputs, &plan); err != nil {
		return nil, err
	}
	if err := ValidatePlan(&plan); err != nil {
		return &ai.SimpleToolResult{ToolContent: fmt.Sprintf(`{"error":%q}`, err.Error())}, nil
	}

	result, err := t.store.ApplyPlan(ctx, &plan)
	if err != nil {
		return &ai.SimpleToolResult{ToolContent: fmt.Sprintf(`{"error":%q}`, err.Error())}, nil
	}

	t.mu.Lock()
	t.lastPlan = clonePlan(&plan)
	t.lastResult = cloneApplyResult(result)
	t.mu.Unlock()

	payload, err := json.Marshal(map[string]any{
		"plan":   plan,
		"result": result,
	})
	if err != nil {
		return nil, err
	}

	return &ai.SimpleToolResult{ToolContent: string(payload)}, nil
}

func (t *ApplyPlanTool) LastPlan() *Plan {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return clonePlan(t.lastPlan)
}

func (t *ApplyPlanTool) LastResult() *ApplyResult {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return cloneApplyResult(t.lastResult)
}

func clonePlan(plan *Plan) *Plan {
	if plan == nil {
		return nil
	}
	companies := append([]Company(nil), plan.Companies...)
	relationships := append([]Relationship(nil), plan.Relationships...)
	sources := append([]string(nil), plan.Sources...)
	return &Plan{
		Keyword:       plan.Keyword,
		Summary:       plan.Summary,
		Companies:     companies,
		Relationships: relationships,
		Sources:       sources,
	}
}

func cloneApplyResult(result *ApplyResult) *ApplyResult {
	if result == nil {
		return nil
	}
	return &ApplyResult{
		CompaniesUpserted:    append([]string(nil), result.CompaniesUpserted...),
		RelationshipsCreated: append([]string(nil), result.RelationshipsCreated...),
	}
}

func decodeInputs(inputs map[string]any, out any) error {
	data, err := json.Marshal(inputs)
	if err != nil {
		return fmt.Errorf("marshal tool input: %w", err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode tool input: %w", err)
	}
	return nil
}

func schemaAsMap(schema any) (map[string]any, error) {
	data, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("marshal schema: %w", err)
	}

	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("unmarshal schema: %w", err)
	}

	return out, nil
}
