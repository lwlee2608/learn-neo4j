package nlquery

import (
	"context"
	"fmt"
	"strings"

	llm "github.com/lwlee2608/learn-neo4j/pkg/ai"
	"github.com/openai/openai-go/v3"
)

const defaultAgentMaxSteps = 4

type AgentConfig struct {
	Model       string
	Temperature float64
	MaxTokens   int
	Provider    *llm.ProviderOption
	Schema      GraphSchema
	MaxSteps    int
}

type QueryAgent struct {
	agent  *llm.Agent
	option llm.CompletionOption
	schema GraphSchema
	tool   *ExecuteCypherTool
}

func NewQueryAgent(completion llm.Completion, executor QueryExecutor, cfg AgentConfig) *QueryAgent {
	if cfg.Schema.Labels == nil {
		cfg.Schema = DefaultGraphSchema()
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = defaultMaxTokens
	}
	if cfg.MaxSteps == 0 {
		cfg.MaxSteps = defaultAgentMaxSteps
	}

	maxTokens := cfg.MaxTokens
	tool := NewExecuteCypherTool(cfg.Schema, executor)

	return &QueryAgent{
		agent:  llm.NewAgent(completion, cfg.MaxSteps),
		schema: cfg.Schema,
		tool:   tool,
		option: llm.CompletionOption{
			Model:       cfg.Model,
			Temperature: cfg.Temperature,
			MaxTokens:   &maxTokens,
			Provider:    cfg.Provider,
		},
	}
}

func (a *QueryAgent) Ask(ctx context.Context, question string) (*Answer, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, fmt.Errorf("question is required")
	}

	result, err := a.agent.Execute(
		ctx,
		[]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(a.systemPrompt()),
			openai.UserMessage(question),
		},
		[]llm.Tool{a.tool},
		a.option,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	answer := &Answer{
		Plan:          a.tool.LastPlan(),
		Result:        a.tool.LastResult(),
		FinalResponse: strings.TrimSpace(result.Content),
	}

	if answer.FinalResponse == "" {
		answer.FinalResponse = "No response generated."
	}

	return answer, nil
}

func (a *QueryAgent) systemPrompt() string {
	return fmt.Sprintf(strings.TrimSpace(`You answer questions about a Neo4j graph.

You have one tool available:
- use it to create a read-only Cypher plan and execute it against Neo4j

Rules:
- Always use the tool before giving a final answer.
- The tool input must include a Cypher plan with query, params, intent, explanation, and read_only.
- The query must be exactly one read-only Cypher statement.
- Use only MATCH, OPTIONAL MATCH, WHERE, WITH, RETURN, ORDER BY, LIMIT, and aggregation when helpful.
- Do not use CREATE, MERGE, DELETE, DETACH, SET, REMOVE, DROP, CALL, APOC, or multi-statement Cypher.
- Use only the provided schema.
- Use parameters for all user-provided values; do not place string literals in the query.
- After the tool returns, answer in natural language using the tool result.
- If no rows are returned, say so clearly.

%s`), a.schema.Prompt())
}
