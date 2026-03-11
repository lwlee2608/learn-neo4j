package neo4j

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Config struct {
	URI      string
	Username string
	Password string
}

type Client struct {
	Driver neo4j.DriverWithContext
}

func NewClient(cfg Config) (*Client, error) {
	driver, err := neo4j.NewDriverWithContext(cfg.URI, neo4j.BasicAuth(cfg.Username, cfg.Password, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to create neo4j driver: %w", err)
	}

	if err := driver.VerifyConnectivity(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to neo4j: %w", err)
	}

	slog.Info("Connected to Neo4j", "uri", cfg.URI)
	return &Client{Driver: driver}, nil
}

func (c *Client) Close(ctx context.Context) error {
	return c.Driver.Close(ctx)
}
