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
	SupplyChain *service.SupplyChainService
}

func SetupRoute(engine *gin.Engine, srvs *Services) {
	engine.Use(middleware.RequestLogger())
	engine.Use(middleware.ErrorHandler())

	healthHandler := handler.NewHealthHandler()
	scHandler := handler.NewSupplyChainHandler(srvs.SupplyChain)

	engine.GET("/health", healthHandler.Check)

	apis := engine.Group("/api/v1")
	{
		apis.POST("/companies", scHandler.CreateCompany)
		apis.GET("/companies", scHandler.ListCompanies)
		apis.GET("/companies/:name", scHandler.GetCompany)

		rels := apis.Group("/relationships")
		{
			rels.POST("/supplies-equipment-to", scHandler.CreateSuppliesEquipmentTo)
			rels.POST("/manufactures-for", scHandler.CreateManufacturesFor)
			rels.POST("/supplies-chips-to", scHandler.CreateSuppliesChipsTo)
			rels.POST("/provides-cloud-for", scHandler.CreateProvidesCloudFor)
			rels.POST("/competes-with", scHandler.CreateCompetesWith)
		}
	}
}
