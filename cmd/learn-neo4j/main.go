package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	internalhttp "github.com/lwlee2608/learn-neo4j/internal/api/http"
	"github.com/lwlee2608/learn-neo4j/internal/repository"
	"github.com/lwlee2608/learn-neo4j/internal/service"
	n "github.com/lwlee2608/learn-neo4j/pkg/neo4j"
)

var AppVersion = "dev"

func main() {
	InitConfig()

	slog.Info("learn-neo4j", "version", AppVersion)

	neo4jClient, err := n.NewClient(config.Neo4j)
	if err != nil {
		slog.Error("Failed to connect to Neo4j", "error", err)
		panic(err)
	}
	defer neo4jClient.Close(context.Background())

	movieRepo := repository.NewMovieRepository(neo4jClient)
	movieSvc := service.NewMovieService(movieRepo)

	services := &internalhttp.Services{
		Movie: movieSvc,
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	engine.Use(gin.Recovery())
	internalhttp.SetupRoute(engine, services)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Http.Port),
		Handler: engine,
	}

	slog.Info("Starting HTTP server", "address", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("HTTP server error", "error", err)
	}
}
