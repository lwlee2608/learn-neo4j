package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/lwlee2608/adder"
	llm "github.com/lwlee2608/learn-neo4j/pkg/ai"
	n "github.com/lwlee2608/learn-neo4j/pkg/neo4j"
)

type EmbeddingConfig struct {
	Model      string
	Dimensions int
}

type Config struct {
	Neo4j      n.Config
	OpenRouter llm.Config
	OpenAI     llm.Config
	Embedding  EmbeddingConfig
}

var config Config

func InitConfig() {
	_ = godotenv.Load()

	adder.SetConfigName("application")
	adder.AddConfigPath(".")
	adder.SetConfigType("yaml")
	adder.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	adder.AutomaticEnv()

	_ = adder.BindEnv("openrouter.apikey", "OPENROUTER_API_KEY")
	_ = adder.BindEnv("openrouter.baseurl", "OPENROUTER_BASE_URL")
	_ = adder.BindEnv("openai.apikey", "OPENAI_API_KEY")
	_ = adder.BindEnv("openai.baseurl", "OPENAI_BASE_URL")

	if err := adder.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := adder.Unmarshal(&config); err != nil {
		panic(err)
	}

	validateConfig(config)
}

func validateConfig(cfg Config) {
	if strings.TrimSpace(cfg.OpenRouter.ApiKey) == "" {
		panic("openrouter.apikey is required (set OPENROUTER_API_KEY env var)")
	}

	if strings.TrimSpace(cfg.OpenRouter.BaseUrl) == "" {
		panic("openrouter.baseurl is required (set OPENROUTER_BASE_URL env var or application.yml)")
	}

	if strings.TrimSpace(cfg.Neo4j.URI) == "" {
		panic("neo4j.uri is required")
	}

	if strings.TrimSpace(cfg.Neo4j.Username) == "" {
		panic("neo4j.username is required")
	}

	if strings.TrimSpace(cfg.Neo4j.Password) == "" {
		panic("neo4j.password is required")
	}
}

func printConfig() {
	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return
	}

	fmt.Println("Config loaded:")
	fmt.Println(string(configJSON))
}
