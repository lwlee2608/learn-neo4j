package graphexpand

import (
	"context"
	"fmt"
	"strings"

	"github.com/lwlee2608/learn-neo4j/pkg/ai"
	"github.com/lwlee2608/learn-neo4j/pkg/exaai"
	"github.com/openai/openai-go/v3"
)

const defaultMaxSteps = 6

type Config struct {
	Model       string
	Temperature float64
	MaxTokens   int
	Provider    *ai.ProviderOption
	MaxSteps    int
}

type Expander struct {
	agent         *ai.Agent
	option        ai.CompletionOption
	applyPlanTool *ApplyPlanTool
	tools         []ai.Tool
}

func NewExpander(completion ai.Completion, store Store, searchClient exaai.Client, cfg Config) *Expander {
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 1200
	}
	if cfg.MaxSteps == 0 {
		cfg.MaxSteps = defaultMaxSteps
	}

	maxTokens := cfg.MaxTokens
	applyTool := NewApplyPlanTool(store)

	return &Expander{
		agent:         ai.NewAgent(completion, cfg.MaxSteps),
		applyPlanTool: applyTool,
		tools: []ai.Tool{
			NewListCompaniesTool(store),
			NewSearchWebTool(searchClient),
			applyTool,
		},
		option: ai.CompletionOption{
			Model:       cfg.Model,
			Temperature: cfg.Temperature,
			MaxTokens:   &maxTokens,
			Provider:    cfg.Provider,
		},
	}
}

func (e *Expander) Expand(ctx context.Context, keyword string) (*Answer, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}

	result, err := e.agent.Execute(
		ctx,
		[]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt()),
			openai.UserMessage(fmt.Sprintf("Expand the Neo4j graph around this keyword: %s", keyword)),
		},
		e.tools,
		e.option,
		nil,
		nil,
	)
	if err != nil {
		return nil, err
	}

	answer := &Answer{
		Plan:          e.applyPlanTool.LastPlan(),
		ApplyResult:   e.applyPlanTool.LastResult(),
		FinalResponse: strings.TrimSpace(result.Content),
	}
	if answer.FinalResponse == "" {
		answer.FinalResponse = "No response generated."
	}
	return answer, nil
}

func systemPrompt() string {
	return strings.TrimSpace(`You expand an AI supply chain Neo4j graph around a keyword.

Available tools:
- list_graph_companies: inspect current companies already in the graph
- search_web: gather evidence from the web
- apply_graph_expansion: create or update companies and relationships in Neo4j

Rules:
- First inspect the existing graph, then do one or more web searches, then apply exactly one expansion plan.
- Only add companies and relationships that have clear support from the search results.
- Use only these relationship types: SUPPLIES_EQUIPMENT_TO, MANUFACTURES_FOR, SUPPLIES_CHIPS_TO, PROVIDES_CLOUD_FOR, COMPETES_WITH.
- The expansion plan must include every company referenced by a relationship.
- Prefer linking new evidence to companies that already exist in the graph.
- Keep the plan compact and high confidence; skip weak or speculative claims.
- After applying the plan, provide a concise final summary describing what was added and why.
- Include a brief note if some likely relationships were skipped due to insufficient evidence.`)
}
