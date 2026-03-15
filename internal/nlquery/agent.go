package nlquery

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/lwlee2608/learn-neo4j/internal/graphschema"
	"github.com/lwlee2608/learn-neo4j/internal/vectorsearch"
	llm "github.com/lwlee2608/learn-neo4j/pkg/ai"
	"github.com/openai/openai-go/v3"
)

//go:embed templates/agent_system_prompt.tmpl
var agentSystemPromptRaw string

var agentSystemPromptTmpl = template.Must(template.New("agent_system").Funcs(template.FuncMap{
	"printf": fmt.Sprintf,
}).Parse(agentSystemPromptRaw))

const defaultAgentMaxSteps = 4

type AgentConfig struct {
	Model        string
	Temperature  float64
	MaxTokens    int
	Provider     *llm.ProviderOption
	Schema       graphschema.GraphSchema
	MaxSteps     int
	VectorSearch *vectorsearch.VectorSearch
}

type QueryAgent struct {
	agent        *llm.Agent
	option       llm.CompletionOption
	schema       graphschema.GraphSchema
	tool         *ExecuteCypherTool
	vectorSearch *vectorsearch.VectorSearch
}

func NewQueryAgent(completion llm.Completion, executor QueryExecutor, cfg AgentConfig) *QueryAgent {
	if cfg.Schema.Labels == nil {
		cfg.Schema = graphschema.Default()
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
		agent:        llm.NewAgent(completion, cfg.MaxSteps),
		schema:       cfg.Schema,
		tool:         tool,
		vectorSearch: cfg.VectorSearch,
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

	var relevantNodes []vectorsearch.SearchResult
	if a.vectorSearch != nil {
		nodes, err := a.vectorSearch.Search(ctx, question, 5)
		if err == nil {
			relevantNodes = nodes
		}
	}

	result, err := a.agent.Execute(
		ctx,
		[]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(a.systemPrompt(relevantNodes)),
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

func (a *QueryAgent) systemPrompt(relevantNodes []vectorsearch.SearchResult) string {
	var buf bytes.Buffer
	agentSystemPromptTmpl.Execute(&buf, struct {
		Schema        string
		RelevantNodes []vectorsearch.SearchResult
	}{
		Schema:        a.schema.Prompt(),
		RelevantNodes: relevantNodes,
	})
	return strings.TrimSpace(buf.String())
}
