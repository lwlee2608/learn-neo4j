package nlquery

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	llm "github.com/lwlee2608/learn-neo4j/pkg/ai"
	"github.com/openai/openai-go/v3"
)

const defaultMaxTokens = 400

type TranslatorConfig struct {
	Model       string
	Temperature float64
	MaxTokens   int
	Provider    *llm.ProviderOption
	Schema      GraphSchema
}

type Translator struct {
	completion llm.Completion
	option     llm.CompletionOption
	schema     GraphSchema
}

func NewTranslator(completion llm.Completion, cfg TranslatorConfig) *Translator {
	if cfg.Schema.Labels == nil {
		cfg.Schema = DefaultGraphSchema()
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = defaultMaxTokens
	}

	maxTokens := cfg.MaxTokens

	return &Translator{
		completion: completion,
		schema:     cfg.Schema,
		option: llm.CompletionOption{
			Model:       cfg.Model,
			Temperature: cfg.Temperature,
			MaxTokens:   &maxTokens,
			Provider:    cfg.Provider,
			ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
				OfJSONObject: &openai.ResponseFormatJSONObjectParam{},
			},
		},
	}
}

func (t *Translator) Translate(ctx context.Context, question string) (*Plan, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, fmt.Errorf("question is required")
	}

	response, err := t.completion.Completions(
		ctx,
		[]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(t.systemPrompt()),
			openai.UserMessage(t.userPrompt(question)),
		},
		nil,
		t.option,
	)
	if err != nil {
		return nil, err
	}

	var plan Plan
	if err := json.Unmarshal([]byte(response.Message.Content), &plan); err != nil {
		return nil, fmt.Errorf("failed to decode LLM response: %w", err)
	}

	if err := ValidatePlan(&plan, t.schema); err != nil {
		return nil, fmt.Errorf("generated plan is invalid: %w", err)
	}

	return &plan, nil
}

func (t *Translator) systemPrompt() string {
	return strings.TrimSpace(`You translate natural language questions into read-only Neo4j Cypher.

Return JSON only with this shape:
{
  "query": "string",
  "params": {"key": "value"},
  "intent": "string",
  "explanation": "string",
  "read_only": true
}

Rules:
- Generate exactly one read-only Cypher statement.
- Use MATCH, OPTIONAL MATCH, WHERE, WITH, RETURN, ORDER BY, LIMIT, and aggregation when helpful.
- Do not use CREATE, MERGE, DELETE, DETACH, SET, REMOVE, DROP, CALL, APOC, or multi-statement Cypher.
- Do not invent labels, relationship types, or properties outside the provided schema.
- Use query parameters for all user-provided values.
- Do not put string literals directly into the query.
- Always set read_only to true.
- If the question cannot be answered with the schema, return a safe fallback query:
  {
    "query": "MATCH (c:Company) RETURN c.name AS message LIMIT 0",
    "params": {},
    "intent": "unsupported",
    "explanation": "The question cannot be answered with the available schema.",
    "read_only": true
  }`)
}

func (t *Translator) userPrompt(question string) string {
	return fmt.Sprintf("%s\n\nUser question:\n%s", t.schema.Prompt(), question)
}
