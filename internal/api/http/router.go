package http

import (
	"github.com/gin-gonic/gin"
	"github.com/lwlee2608/learn-neo4j/internal/api/http/handler"
	"github.com/lwlee2608/learn-neo4j/internal/api/http/middleware"
	"github.com/lwlee2608/learn-neo4j/internal/service"
)

type Config struct {
	Port uint
}

type Services struct {
	Movie *service.MovieService
}

func SetupRoute(engine *gin.Engine, srvs *Services) {
	engine.Use(middleware.RequestLogger())
	engine.Use(middleware.ErrorHandler())

	healthHandler := handler.NewHealthHandler()
	movieHandler := handler.NewMovieHandler(srvs.Movie)

	engine.GET("/health", healthHandler.Check)

	apis := engine.Group("/api/v1")
	{
		apis.POST("/movies", movieHandler.CreateMovie)
		apis.GET("/movies", movieHandler.ListMovies)
		apis.GET("/movies/:title", movieHandler.GetMovie)

		apis.POST("/persons", movieHandler.CreatePerson)
		apis.GET("/persons", movieHandler.ListPersons)

		apis.POST("/acted-in", movieHandler.CreateActedIn)
	}
}
