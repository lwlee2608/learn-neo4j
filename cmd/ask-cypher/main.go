package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lwlee2608/learn-neo4j/internal/nlquery"
	llm "github.com/lwlee2608/learn-neo4j/pkg/ai"
	n "github.com/lwlee2608/learn-neo4j/pkg/neo4j"
)

func main() {
	InitConfig()

	questionFlag := flag.String("question", "", "natural language question to translate")
	modelFlag := flag.String("model", envOrDefault("OPENROUTER_MODEL", "openai/gpt-4.1-mini"), "OpenRouter model to use")
	temperatureFlag := flag.Float64("temperature", 0, "sampling temperature")
	maxTokensFlag := flag.Int("max-tokens", 400, "maximum completion tokens")
	executeFlag := flag.Bool("execute", false, "execute the validated query against Neo4j")
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

	completion := llm.NewOpenAIService(config.OpenRouter.ApiKey, config.OpenRouter.BaseUrl)
	translator := nlquery.NewTranslator(completion, nlquery.TranslatorConfig{
		Model:       *modelFlag,
		Temperature: *temperatureFlag,
		MaxTokens:   *maxTokensFlag,
	})

	ctx := context.Background()
	plan, err := translator.Translate(ctx, question)
	if err != nil {
		log.Fatalf("translate question: %v", err)
	}

	printJSON("plan", plan)

	if !*executeFlag {
		return
	}

	client, err := n.NewClient(config.Neo4j)
	if err != nil {
		log.Fatalf("connect to neo4j: %v", err)
	}
	defer client.Close(ctx)

	result, err := nlquery.ExecuteReadOnly(ctx, client, plan)
	if err != nil {
		log.Fatalf("execute plan: %v", err)
	}

	printJSON("result", result)
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
