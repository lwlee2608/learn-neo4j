package ai

import (
	"context"
	"errors"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/pagination"
)

var (
	_ Completion = (*OpenAI)(nil)
	_ Embedding  = (*OpenAI)(nil)
)

type Config struct {
	ApiKey  string `mask:"first=3,last=4"`
	BaseUrl string
}

type OpenAI struct {
	client openai.Client
}

func NewOpenAIService(apiKey string, baseUrl string) *OpenAI {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseUrl),
	)
	return &OpenAI{
		client: client,
	}
}

func (s *OpenAI) ParamsCompletions(ctx context.Context, params openai.ChatCompletionNewParams, provider *ProviderOption) (CompletionResponse, error) {
	var err error
	var completion *openai.ChatCompletion

	if provider != nil {
		completion, err = s.client.Chat.Completions.New(
			ctx,
			params,
			option.WithJSONSet("provider",
				map[string]any{
					"allow_fallbacks": provider.AllowFallbacks,
					"data_collection": "deny",
					"sort":            provider.Sort,
				}),
		)
	} else {
		completion, err = s.client.Chat.Completions.New(ctx, params)
	}

	if err != nil {
		return CompletionResponse{}, err
	}
	if len(completion.Choices) == 0 {
		return CompletionResponse{}, errors.New("no completion choices returned from OpenAI API")
	}
	return CompletionResponse{
		Message: completion.Choices[0].Message,
		Usage:   completion.Usage,
	}, nil
}

func (s *OpenAI) Completions(
	ctx context.Context,
	messages []openai.ChatCompletionMessageParamUnion,
	tools []openai.ChatCompletionToolUnionParam,
	option CompletionOption,
) (CompletionResponse, error) {
	params := openai.ChatCompletionNewParams{
		Messages:       messages,
		Model:          option.Model,
		Temperature:    openai.Float(option.Temperature),
		ResponseFormat: option.ResponseFormat,
	}

	if option.MaxTokens != nil {
		params.MaxTokens = openai.Int(int64(*option.MaxTokens))
	}

	if len(tools) > 0 {
		params.Tools = tools
	}

	return s.ParamsCompletions(ctx, params, option.Provider)
}

func (s *OpenAI) ParamsCompletionsStream(ctx context.Context, params openai.ChatCompletionNewParams, provider *ProviderOption, callback StreamCallback) error {
	var stream interface {
		Next() bool
		Current() openai.ChatCompletionChunk
		Err() error
		Close() error
	}

	if provider != nil {
		stream = s.client.Chat.Completions.NewStreaming(
			ctx,
			params,
			option.WithJSONSet("provider",
				map[string]any{
					"allow_fallbacks": provider.AllowFallbacks,
					"data_collection": "deny",
					"sort":            provider.Sort,
				}),
		)
	} else {
		stream = s.client.Chat.Completions.NewStreaming(ctx, params)
	}
	defer stream.Close()

	toolCallsMap := make(map[int]*openai.ChatCompletionMessageToolCallUnion)

	for stream.Next() {
		chunk := stream.Current()

		if len(chunk.Choices) == 0 {
			continue
		}

		choice := chunk.Choices[0]
		delta := choice.Delta

		streamChunk := StreamChunk{
			Content:      delta.Content,
			Role:         delta.Role,
			FinishReason: choice.FinishReason,
			Done:         false,
		}

		if len(delta.ToolCalls) > 0 {
			for _, toolCall := range delta.ToolCalls {
				index := int(toolCall.Index)

				if _, exists := toolCallsMap[index]; !exists {
					toolCallsMap[index] = &openai.ChatCompletionMessageToolCallUnion{
						ID:   toolCall.ID,
						Type: toolCall.Type,
						Function: openai.ChatCompletionMessageFunctionToolCallFunction{
							Name:      toolCall.Function.Name,
							Arguments: "",
						},
					}
				}

				if toolCall.Function.Arguments != "" {
					tc := toolCallsMap[index]
					tc.Function.Arguments += toolCall.Function.Arguments
				}
			}

			var toolCalls []openai.ChatCompletionMessageToolCallUnion
			for _, tc := range toolCallsMap {
				toolCalls = append(toolCalls, *tc)
			}
			streamChunk.ToolCalls = toolCalls
		}

		if err := callback(streamChunk); err != nil {
			return err
		}
	}

	if err := stream.Err(); err != nil {
		return err
	}

	finalChunk := StreamChunk{
		Done: true,
	}

	if err := callback(finalChunk); err != nil {
		return err
	}

	return nil
}

func (s *OpenAI) CompletionsStream(
	ctx context.Context,
	messages []openai.ChatCompletionMessageParamUnion,
	tools []openai.ChatCompletionToolUnionParam,
	option CompletionOption,
	callback StreamCallback,
) error {
	params := openai.ChatCompletionNewParams{
		Messages:       messages,
		Model:          option.Model,
		Temperature:    openai.Float(option.Temperature),
		ResponseFormat: option.ResponseFormat,
	}

	if option.MaxTokens != nil {
		params.MaxTokens = openai.Int(int64(*option.MaxTokens))
	}

	if len(tools) > 0 {
		params.Tools = tools
	}

	return s.ParamsCompletionsStream(ctx, params, option.Provider, callback)
}

func (s *OpenAI) Models(ctx context.Context) (*pagination.Page[openai.Model], error) {
	return s.client.Models.List(ctx)
}

func (s *OpenAI) Embedding(ctx context.Context, input string, model string) ([]float32, error) {
	embedding, err := s.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Model: openai.EmbeddingModel(model),
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(input),
		},
	})
	if err != nil {
		return nil, err
	}

	if len(embedding.Data) == 0 {
		return nil, errors.New("no embedding data returned from OpenAI")
	}

	f64 := embedding.Data[0].Embedding
	f32 := make([]float32, len(f64))
	for i, v := range f64 {
		f32[i] = float32(v)
	}

	return f32, nil
}
