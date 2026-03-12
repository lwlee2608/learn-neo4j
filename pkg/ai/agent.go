package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
)

type ToolResult interface {
	Content() string
	Error() error
}

type SimpleToolResult struct {
	ToolContent string
	ToolError   error
}

func (t *SimpleToolResult) Content() string {
	return t.ToolContent
}

func (t *SimpleToolResult) Error() error {
	return t.ToolError
}

type ToolName string

type Tool interface {
	Definition() openai.ChatCompletionToolUnionParam
	Execute(ctx context.Context, inputs map[string]any) (ToolResult, error)
}

type AgentResult struct {
	Content     string
	ToolCalls   []openai.ChatCompletionMessageToolCallUnion
	ToolResults []ToolResult
	Usage       openai.CompletionUsage
}

type Agent struct {
	aiService Completion
	maxSteps  int
}

func NewAgent(aiService Completion, maxSteps int) *Agent {
	return &Agent{
		aiService: aiService,
		maxSteps:  maxSteps,
	}
}

type PreTool struct {
	Content string `json:"content"`
	Name    string `json:"name"`
	Args    string `json:"args"`
}

type PreToolCallback func(PreTool)

type PostTool struct {
	Result ToolResult `json:"result"`
}

type PosetToolCallback func(PostTool)

func (a *Agent) Execute(
	ctx context.Context,
	messages []openai.ChatCompletionMessageParamUnion,
	tools []Tool,
	option CompletionOption,
	preCallback PreToolCallback,
	postCallback PosetToolCallback,
) (*AgentResult, error) {

	toolCalls := make([]openai.ChatCompletionMessageToolCallUnion, 0)
	toolResults := make([]ToolResult, 0)
	responseContent := ""
	totalUsage := openai.CompletionUsage{}

	toolDefinitions := make([]openai.ChatCompletionToolUnionParam, 0)
	toolsMap := make(map[string]Tool, 0)
	for _, tool := range tools {
		function := tool.Definition().GetFunction()
		if function != nil {
			toolsMap[function.Name] = tool
		}
		toolDefinitions = append(toolDefinitions, tool.Definition())
	}

	currentStep := 0
	for currentStep < a.maxSteps {

		completion, err := a.aiService.Completions(
			ctx,
			messages,
			toolDefinitions,
			option,
		)
		if err != nil {
			return nil, err
		}

		messages = append(messages, completion.Message.ToParam())

		// Accumulate usage from this completion
		totalUsage.PromptTokens += completion.Usage.PromptTokens
		totalUsage.CompletionTokens += completion.Usage.CompletionTokens
		totalUsage.TotalTokens += completion.Usage.TotalTokens

		if len(completion.Message.ToolCalls) == 0 {
			return &AgentResult{
				Content:     completion.Message.Content,
				ToolCalls:   toolCalls,   // Return accumulated tool calls
				ToolResults: toolResults, // Return accumulated tool results
				Usage:       totalUsage,
			}, nil
		}

		for _, toolCall := range completion.Message.ToolCalls {

			// Stsream Callback
			if preCallback != nil {
				preCallback(PreTool{
					Content: completion.Message.Content,
					Name:    toolCall.Function.Name,
					Args:    toolCall.Function.Arguments,
				})
			}

			tool, ok := toolsMap[toolCall.Function.Name]
			if !ok {
				toolResult := &SimpleToolResult{
					ToolContent: fmt.Sprintf("Tool %s not found", toolCall.Function.Name),
				}
				messages = append(messages, openai.ToolMessage(toolResult.Content(), toolCall.ID))
				toolCalls = append(toolCalls, toolCall)
				toolResults = append(toolResults, toolResult)
				continue
			}

			var args map[string]any
			err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
			if err != nil {
				toolResult := &SimpleToolResult{
					ToolContent: fmt.Sprintf("Failed to unmarshal arguments for tool %s: %v", toolCall.Function.Name, err),
					ToolError:   err,
				}
				messages = append(messages, openai.ToolMessage(toolResult.Content(), toolCall.ID))
				toolCalls = append(toolCalls, toolCall)
				toolResults = append(toolResults, toolResult)
				continue
			}

			toolResult, err := tool.Execute(ctx, args)
			if err != nil {
				toolResult = &SimpleToolResult{
					ToolContent: fmt.Sprintf("Error executing tool %s: %v", toolCall.Function.Name, err),
					ToolError:   err,
				}
			}

			if postCallback != nil {
				postCallback(PostTool{
					Result: toolResult,
				})
			}

			messages = append(messages, openai.ToolMessage(toolResult.Content(), toolCall.ID))
			toolCalls = append(toolCalls, toolCall)
			toolResults = append(toolResults, toolResult)
		}

		responseContent = completion.Message.Content
		currentStep++
	}

	return &AgentResult{
		Content:     responseContent,
		ToolCalls:   toolCalls,
		ToolResults: toolResults,
		Usage:       totalUsage,
	}, nil
}

// ExecuteStream performs the agent loop with streaming support
func (a *Agent) ExecuteStream(
	ctx context.Context,
	messages []openai.ChatCompletionMessageParamUnion,
	tools []Tool,
	option CompletionOption,
	streamCallback StreamCallback,
	preCallback PreToolCallback,
	postCallback PosetToolCallback,
) (*AgentResult, error) {

	toolCalls := make([]openai.ChatCompletionMessageToolCallUnion, 0)
	toolResults := make([]ToolResult, 0)
	responseContent := ""
	totalUsage := openai.CompletionUsage{}

	toolDefinitions := make([]openai.ChatCompletionToolUnionParam, 0)
	toolsMap := make(map[string]Tool, 0)
	for _, tool := range tools {
		function := tool.Definition().GetFunction()
		if function != nil {
			toolsMap[function.Name] = tool
		}
		toolDefinitions = append(toolDefinitions, tool.Definition())
	}

	currentStep := 0
	for currentStep < a.maxSteps {
		var currentContent string
		var currentToolCalls []openai.ChatCompletionMessageToolCallUnion
		var currentUsage openai.CompletionUsage

		// Wrap the stream callback to accumulate content and tool calls
		err := a.aiService.CompletionsStream(
			ctx,
			messages,
			toolDefinitions,
			option,
			func(chunk StreamChunk) error {
				// Accumulate content
				if chunk.Content != "" {
					currentContent += chunk.Content
				}

				// Accumulate tool calls
				if len(chunk.ToolCalls) > 0 {
					currentToolCalls = chunk.ToolCalls
				}

				// Capture usage
				if chunk.Done && chunk.Usage != nil {
					currentUsage = *chunk.Usage
				}

				// Forward the chunk to the user's callback, but don't forward Done=true
				// from intermediate completions (only the final one)
				if chunk.Done {
					// Don't forward Done for intermediate completions
					return nil
				}
				return streamCallback(chunk)
			},
		)

		if err != nil {
			return nil, err
		}

		// Accumulate total usage
		totalUsage.PromptTokens += currentUsage.PromptTokens
		totalUsage.CompletionTokens += currentUsage.CompletionTokens
		totalUsage.TotalTokens += currentUsage.TotalTokens

		// Build the assistant message with content and tool calls
		assistantMessage := openai.ChatCompletionMessage{
			Role:      "assistant",
			Content:   currentContent,
			ToolCalls: currentToolCalls,
		}

		messages = append(messages, assistantMessage.ToParam())

		// If no tool calls, we're done
		if len(currentToolCalls) == 0 {
			// Send final Done chunk
			streamCallback(StreamChunk{
				Done:  true,
				Usage: &totalUsage,
			})
			return &AgentResult{
				Content:     currentContent,
				ToolCalls:   toolCalls,
				ToolResults: toolResults,
				Usage:       totalUsage,
			}, nil
		}

		// Execute tool calls
		for _, toolCall := range currentToolCalls {
			// Pre-tool callback
			if preCallback != nil {
				preCallback(PreTool{
					Content: currentContent,
					Name:    toolCall.Function.Name,
					Args:    toolCall.Function.Arguments,
				})
			}

			tool, ok := toolsMap[toolCall.Function.Name]
			if !ok {
				toolResult := &SimpleToolResult{
					ToolContent: fmt.Sprintf("Tool %s not found", toolCall.Function.Name),
				}
				messages = append(messages, openai.ToolMessage(toolResult.Content(), toolCall.ID))
				toolCalls = append(toolCalls, toolCall)
				toolResults = append(toolResults, toolResult)
				continue
			}

			var args map[string]any
			err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
			if err != nil {
				toolResult := &SimpleToolResult{
					ToolContent: fmt.Sprintf("Failed to unmarshal arguments for tool %s: %v", toolCall.Function.Name, err),
					ToolError:   err,
				}
				messages = append(messages, openai.ToolMessage(toolResult.Content(), toolCall.ID))
				toolCalls = append(toolCalls, toolCall)
				toolResults = append(toolResults, toolResult)
				continue
			}

			toolResult, err := tool.Execute(ctx, args)
			if err != nil {
				toolResult = &SimpleToolResult{
					ToolContent: fmt.Sprintf("Error executing tool %s: %v", toolCall.Function.Name, err),
					ToolError:   err,
				}
			}

			if postCallback != nil {
				postCallback(PostTool{
					Result: toolResult,
				})
			}

			// Stream the tool result
			streamCallback(StreamChunk{
				ToolResults: []StreamToolResult{
					{
						Name:    toolCall.Function.Name,
						Content: toolResult.Content(),
						Error:   toolResult.Error(),
					},
				},
			})

			messages = append(messages, openai.ToolMessage(toolResult.Content(), toolCall.ID))
			toolCalls = append(toolCalls, toolCall)
			toolResults = append(toolResults, toolResult)
		}

		responseContent = currentContent
		currentStep++
	}

	// Send final Done chunk (reached max steps)
	streamCallback(StreamChunk{
		Done:  true,
		Usage: &totalUsage,
	})

	return &AgentResult{
		Content:     responseContent,
		ToolCalls:   toolCalls,
		ToolResults: toolResults,
		Usage:       totalUsage,
	}, nil
}
