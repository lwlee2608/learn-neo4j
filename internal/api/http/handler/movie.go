package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lwlee2608/learn-neo4j/internal/domain"
	"github.com/lwlee2608/learn-neo4j/internal/service"
)

type MovieHandler struct {
	svc *service.MovieService
}

func NewMovieHandler(svc *service.MovieService) *MovieHandler {
	return &MovieHandler{svc: svc}
}

func (h *MovieHandler) CreateMovie(ctx *gin.Context) {
	var movie domain.Movie
	if err := ctx.ShouldBindJSON(&movie); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CreateMovie(ctx.Request.Context(), movie); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, movie)
}

func (h *MovieHandler) ListMovies(ctx *gin.Context) {
	movies, err := h.svc.ListMovies(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, movies)
}

func (h *MovieHandler) GetMovie(ctx *gin.Context) {
	title := ctx.Param("title")
	movie, err := h.svc.GetMovie(ctx.Request.Context(), title)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, movie)
}

func (h *MovieHandler) CreatePerson(ctx *gin.Context) {
	var person domain.Person
	if err := ctx.ShouldBindJSON(&person); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CreatePerson(ctx.Request.Context(), person); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, person)
}

func (h *MovieHandler) ListPersons(ctx *gin.Context) {
	persons, err := h.svc.ListPersons(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, persons)
}

func (h *MovieHandler) CreateActedIn(ctx *gin.Context) {
	var actedIn domain.ActedIn
	if err := ctx.ShouldBindJSON(&actedIn); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.CreateActedIn(ctx.Request.Context(), actedIn); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, actedIn)
}
