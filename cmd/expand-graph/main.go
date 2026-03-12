package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lwlee2608/learn-neo4j/internal/graphexpand"
	llm "github.com/lwlee2608/learn-neo4j/pkg/ai"
	"github.com/lwlee2608/learn-neo4j/pkg/exaai"
	n "github.com/lwlee2608/learn-neo4j/pkg/neo4j"
)

func main() {
	InitConfig()

	keywordFlag := flag.String("keyword", "", "keyword or company name to expand around")
	modelFlag := flag.String("model", envOrDefault("OPENROUTER_MODEL", "openai/gpt-4.1-mini"), "OpenRouter model to use")
	temperatureFlag := flag.Float64("temperature", 0.2, "sampling temperature")
	maxTokensFlag := flag.Int("max-tokens", 1200, "maximum completion tokens")
	flag.Parse()

	keyword := strings.TrimSpace(*keywordFlag)
	if keyword == "" {
		keyword = strings.TrimSpace(strings.Join(flag.Args(), " "))
	}
	if keyword == "" {
		log.Fatal("keyword is required; pass -keyword or trailing args")
	}

	ctx := context.Background()
	neo4jClient, err := n.NewClient(config.Neo4j)
	if err != nil {
		log.Fatalf("connect to neo4j: %v", err)
	}
	defer neo4jClient.Close(ctx)

	completion := llm.NewOpenAIService(config.OpenRouter.ApiKey, config.OpenRouter.BaseUrl)
	searchClient := exaai.NewClient(config.Exa.ApiKey)
	store := graphexpand.NewNeo4jStore(neo4jClient)
	expander := graphexpand.NewExpander(completion, store, searchClient, graphexpand.Config{
		Model:       *modelFlag,
		Temperature: *temperatureFlag,
		MaxTokens:   *maxTokensFlag,
	})

	answer, err := expander.Expand(ctx, keyword)
	if err != nil {
		log.Fatalf("expand graph: %v", err)
	}

	if answer.Plan != nil {
		printJSON("plan", answer.Plan)
	}
	if answer.ApplyResult != nil {
		printJSON("apply_result", answer.ApplyResult)
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
