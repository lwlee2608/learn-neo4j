package main

import (
	"strings"

	"github.com/joho/godotenv"
	"github.com/lwlee2608/adder"
	llm "github.com/lwlee2608/learn-neo4j/pkg/ai"
	"github.com/lwlee2608/learn-neo4j/pkg/exaai"
	n "github.com/lwlee2608/learn-neo4j/pkg/neo4j"
)

type Config struct {
	Neo4j      n.Config
	OpenRouter llm.Config
	Exa        exaai.Config
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
	_ = adder.BindEnv("exa.apikey", "EXAAI_API_KEY")

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
	if strings.TrimSpace(cfg.Exa.ApiKey) == "" {
		panic("exa.apikey is required (set EXAAI_API_KEY env var)")
	}
	if strings.TrimSpace(cfg.Neo4j.URI) == "" || strings.TrimSpace(cfg.Neo4j.Username) == "" || strings.TrimSpace(cfg.Neo4j.Password) == "" {
		panic("neo4j config is required")
	}
}
