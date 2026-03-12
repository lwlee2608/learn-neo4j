package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lwlee2608/learn-neo4j/internal/domain"
	"github.com/lwlee2608/learn-neo4j/internal/service"
)

type SupplyChainHandler struct {
	svc *service.SupplyChainService
}

func NewSupplyChainHandler(svc *service.SupplyChainService) *SupplyChainHandler {
	return &SupplyChainHandler{svc: svc}
}

func (h *SupplyChainHandler) CreateCompany(ctx *gin.Context) {
	var company domain.Company
	if err := ctx.ShouldBindJSON(&company); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CreateCompany(ctx.Request.Context(), company); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, company)
}

func (h *SupplyChainHandler) ListCompanies(ctx *gin.Context) {
	companies, err := h.svc.ListCompanies(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, companies)
}

func (h *SupplyChainHandler) GetCompany(ctx *gin.Context) {
	name := ctx.Param("name")
	company, err := h.svc.GetCompany(ctx.Request.Context(), name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, company)
}

func (h *SupplyChainHandler) CreateChip(ctx *gin.Context) {
	var chip domain.Chip
	if err := ctx.ShouldBindJSON(&chip); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CreateChip(ctx.Request.Context(), chip); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, chip)
}

func (h *SupplyChainHandler) ListChips(ctx *gin.Context) {
	chips, err := h.svc.ListChips(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, chips)
}

func (h *SupplyChainHandler) GetChip(ctx *gin.Context) {
	name := ctx.Param("name")
	chip, err := h.svc.GetChip(ctx.Request.Context(), name)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, chip)
}

func (h *SupplyChainHandler) CreateDesigned(ctx *gin.Context) {
	var rel domain.Designed
	if err := ctx.ShouldBindJSON(&rel); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CreateDesigned(ctx.Request.Context(), rel); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, rel)
}

func (h *SupplyChainHandler) CreateManufactures(ctx *gin.Context) {
	var rel domain.Manufactures
	if err := ctx.ShouldBindJSON(&rel); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CreateManufactures(ctx.Request.Context(), rel); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, rel)
}

func (h *SupplyChainHandler) CreateSuppliesEquipmentTo(ctx *gin.Context) {
	var rel domain.SuppliesEquipmentTo
	if err := ctx.ShouldBindJSON(&rel); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CreateSuppliesEquipmentTo(ctx.Request.Context(), rel); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, rel)
}

func (h *SupplyChainHandler) CreateProvidesCloudFor(ctx *gin.Context) {
	var rel domain.ProvidesCloudFor
	if err := ctx.ShouldBindJSON(&rel); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CreateProvidesCloudFor(ctx.Request.Context(), rel); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, rel)
}

func (h *SupplyChainHandler) CreateUses(ctx *gin.Context) {
	var rel domain.Uses
	if err := ctx.ShouldBindJSON(&rel); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CreateUses(ctx.Request.Context(), rel); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, rel)
}
