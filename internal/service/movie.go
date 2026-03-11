package service

import (
	"context"

	"github.com/lwlee2608/learn-neo4j/internal/domain"
	"github.com/lwlee2608/learn-neo4j/internal/repository"
)

type MovieService struct {
	repo *repository.MovieRepository
}

func NewMovieService(repo *repository.MovieRepository) *MovieService {
	return &MovieService{repo: repo}
}

func (s *MovieService) CreateMovie(ctx context.Context, movie domain.Movie) error {
	return s.repo.CreateMovie(ctx, movie)
}

func (s *MovieService) ListMovies(ctx context.Context) ([]domain.Movie, error) {
	return s.repo.ListMovies(ctx)
}

func (s *MovieService) GetMovie(ctx context.Context, title string) (*domain.MovieWithCast, error) {
	return s.repo.GetMovie(ctx, title)
}

func (s *MovieService) CreatePerson(ctx context.Context, person domain.Person) error {
	return s.repo.CreatePerson(ctx, person)
}

func (s *MovieService) ListPersons(ctx context.Context) ([]domain.Person, error) {
	return s.repo.ListPersons(ctx)
}

func (s *MovieService) CreateActedIn(ctx context.Context, actedIn domain.ActedIn) error {
	return s.repo.CreateActedIn(ctx, actedIn)
}
