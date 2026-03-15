package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lwlee2608/learn-neo4j/internal/graphschema"
	"github.com/lwlee2608/learn-neo4j/internal/nlquery"
	llm "github.com/lwlee2608/learn-neo4j/pkg/ai"
	n "github.com/lwlee2608/learn-neo4j/pkg/neo4j"
)

func main() {
	InitConfig()

	questionFlag := flag.String("question", "", "natural language question to translate")
	modelFlag := flag.String("model", envOrDefault("OPENROUTER_MODEL", "anthropic/claude-sonnet-4-6"), "OpenRouter model to use")
	temperatureFlag := flag.Float64("temperature", 0, "sampling temperature")
	maxTokensFlag := flag.Int("max-tokens", 4096, "maximum completion tokens")
	printConfigFlag := flag.Bool("print-config", false, "print resolved config before running")
	flag.Parse()

	question := strings.TrimSpace(*questionFlag)
	if question == "" {
		question = strings.TrimSpace(strings.Join(flag.Args(), " "))
	}
	if question == "" {
		log.Fatal("question is required; pass -question or trailing args")
	}

	if *printConfigFlag {
		printConfig()
	}

	client, err := n.NewClient(config.Neo4j)
	if err != nil {
		log.Fatalf("connect to neo4j: %v", err)
	}
	defer client.Close(context.Background())

	completion := llm.NewOpenAIService(config.OpenRouter.ApiKey, config.OpenRouter.BaseUrl)
	executor := nlquery.NewNeo4jExecutor(client, graphschema.Default())
	queryAgent := nlquery.NewQueryAgent(completion, executor, nlquery.AgentConfig{
		Model:       *modelFlag,
		Temperature: *temperatureFlag,
		MaxTokens:   *maxTokensFlag,
	})

	ctx := context.Background()
	answer, err := queryAgent.Ask(ctx, question)
	if err != nil {
		log.Fatalf("ask question: %v", err)
	}

	if answer.Plan != nil {
		printJSON("plan", answer.Plan)
	}
	if answer.Result != nil {
		printJSON("result", answer.Result)
	}
	printText("answer", answer.FinalResponse)
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func printJSON(label string, value any) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		log.Fatalf("marshal %s: %v", label, err)
	}

	fmt.Printf("%s:\n%s\n", strings.ToUpper(label), data)
}

func printText(label string, value string) {
	fmt.Printf("%s:\n%s\n", strings.ToUpper(label), value)
}
