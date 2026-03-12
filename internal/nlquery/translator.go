package nlquery

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	llm "github.com/lwlee2608/learn-neo4j/pkg/ai"
	"github.com/openai/openai-go/v3"
)

//go:embed templates/translator_system_prompt.tmpl
var translatorSystemPrompt string

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
	return strings.TrimSpace(translatorSystemPrompt)
}

func (t *Translator) userPrompt(question string) string {
	return fmt.Sprintf("%s\n\nUser question:\n%s", t.schema.Prompt(), question)
}
