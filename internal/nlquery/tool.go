package nlquery

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	llm "github.com/lwlee2608/learn-neo4j/pkg/ai"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
)

type QueryExecutor interface {
	ExecuteReadOnly(ctx context.Context, plan *Plan) (*QueryResult, error)
}

type executeCypherInput struct {
	Query       string         `json:"query"`
	Params      map[string]any `json:"params"`
	Intent      string         `json:"intent"`
	Explanation string         `json:"explanation"`
	ReadOnly    bool           `json:"read_only"`
}

type ExecuteCypherTool struct {
	schema   GraphSchema
	executor QueryExecutor

	mu         sync.RWMutex
	lastPlan   *Plan
	lastResult *QueryResult
}

func NewExecuteCypherTool(schema GraphSchema, executor QueryExecutor) *ExecuteCypherTool {
	return &ExecuteCypherTool{
		schema:   schema,
		executor: executor,
	}
}

func (t *ExecuteCypherTool) Definition() openai.ChatCompletionToolUnionParam {
	parameters, err := schemaAsMap(llm.GenerateSchema[executeCypherInput]())
	if err != nil {
		panic(fmt.Sprintf("generate execute_cypher_query schema: %v", err))
	}

	return openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
		Name:        "execute_cypher_query",
		Description: param.NewOpt("Validate and execute a read-only Cypher query against the Neo4j graph"),
		Parameters:  parameters,
	})
}

func (t *ExecuteCypherTool) Execute(ctx context.Context, inputs map[string]any) (llm.ToolResult, error) {
	data, err := json.Marshal(inputs)
	if err != nil {
		return nil, fmt.Errorf("marshal tool input: %w", err)
	}

	var input executeCypherInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("decode tool input: %w", err)
	}

	plan := &Plan{
		Query:       input.Query,
		Params:      input.Params,
		Intent:      input.Intent,
		Explanation: input.Explanation,
		ReadOnly:    input.ReadOnly,
	}

	if err := ValidatePlan(plan, t.schema); err != nil {
		return &llm.SimpleToolResult{ToolContent: fmt.Sprintf(`{"error":%q}`, err.Error())}, nil
	}

	result, err := t.executor.ExecuteReadOnly(ctx, plan)
	if err != nil {
		return &llm.SimpleToolResult{ToolContent: fmt.Sprintf(`{"error":%q}`, err.Error())}, nil
	}

	t.mu.Lock()
	t.lastPlan = clonePlan(plan)
	t.lastResult = cloneResult(result)
	t.mu.Unlock()

	payload := map[string]any{
		"plan":   plan,
		"result": result,
	}

	response, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal tool output: %w", err)
	}

	return &llm.SimpleToolResult{ToolContent: string(response)}, nil
}

func (t *ExecuteCypherTool) LastPlan() *Plan {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return clonePlan(t.lastPlan)
}

func (t *ExecuteCypherTool) LastResult() *QueryResult {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return cloneResult(t.lastResult)
}

func clonePlan(plan *Plan) *Plan {
	if plan == nil {
		return nil
	}

	params := make(map[string]any, len(plan.Params))
	for key, value := range plan.Params {
		params[key] = value
	}

	return &Plan{
		Query:       plan.Query,
		Params:      params,
		Intent:      plan.Intent,
		Explanation: plan.Explanation,
		ReadOnly:    plan.ReadOnly,
	}
}

func cloneResult(result *QueryResult) *QueryResult {
	if result == nil {
		return nil
	}

	records := make([]map[string]any, 0, len(result.Records))
	for _, record := range result.Records {
		cloned := make(map[string]any, len(record))
		for key, value := range record {
			cloned[key] = value
		}
		records = append(records, cloned)
	}

	return &QueryResult{
		Records: records,
		Count:   result.Count,
	}
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
