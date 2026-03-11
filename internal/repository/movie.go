package repository

import (
	"context"
	"fmt"

	"github.com/lwlee2608/learn-neo4j/internal/domain"
	n "github.com/lwlee2608/learn-neo4j/pkg/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type MovieRepository struct {
	client *n.Client
}

func NewMovieRepository(client *n.Client) *MovieRepository {
	return &MovieRepository{client: client}
}

func (r *MovieRepository) CreateMovie(ctx context.Context, movie domain.Movie) error {
	_, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		"CREATE (m:Movie {title: $title, released: $released, tagline: $tagline})",
		map[string]any{
			"title":    movie.Title,
			"released": movie.Released,
			"tagline":  movie.Tagline,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	return err
}

func (r *MovieRepository) ListMovies(ctx context.Context) ([]domain.Movie, error) {
	result, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		"MATCH (m:Movie) RETURN m.title AS title, m.released AS released, m.tagline AS tagline ORDER BY m.released",
		nil,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	if err != nil {
		return nil, err
	}

	var movies []domain.Movie
	for _, record := range result.Records {
		title, _ := record.Get("title")
		released, _ := record.Get("released")
		tagline, _ := record.Get("tagline")
		movies = append(movies, domain.Movie{
			Title:    title.(string),
			Released: int(released.(int64)),
			Tagline:  stringOrEmpty(tagline),
		})
	}
	return movies, nil
}

func (r *MovieRepository) GetMovie(ctx context.Context, title string) (*domain.MovieWithCast, error) {
	result, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		`MATCH (m:Movie {title: $title})
		 OPTIONAL MATCH (p:Person)-[r:ACTED_IN]->(m)
		 RETURN m.title AS title, m.released AS released, m.tagline AS tagline,
		        collect({name: p.name, role: r.role}) AS cast`,
		map[string]any{"title": title},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	if err != nil {
		return nil, err
	}

	if len(result.Records) == 0 {
		return nil, fmt.Errorf("movie not found: %s", title)
	}

	record := result.Records[0]
	movieTitle, _ := record.Get("title")
	released, _ := record.Get("released")
	tagline, _ := record.Get("tagline")
	castRaw, _ := record.Get("cast")

	movie := domain.Movie{
		Title:    movieTitle.(string),
		Released: int(released.(int64)),
		Tagline:  stringOrEmpty(tagline),
	}

	var cast []domain.CastMember
	for _, c := range castRaw.([]any) {
		m := c.(map[string]any)
		if m["name"] == nil {
			continue
		}
		cast = append(cast, domain.CastMember{
			Name: m["name"].(string),
			Role: stringOrEmpty(m["role"]),
		})
	}

	return &domain.MovieWithCast{Movie: movie, Cast: cast}, nil
}

func (r *MovieRepository) CreatePerson(ctx context.Context, person domain.Person) error {
	_, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		"CREATE (p:Person {name: $name, born: $born})",
		map[string]any{
			"name": person.Name,
			"born": person.Born,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	return err
}

func (r *MovieRepository) ListPersons(ctx context.Context) ([]domain.Person, error) {
	result, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		"MATCH (p:Person) RETURN p.name AS name, p.born AS born ORDER BY p.name",
		nil,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	if err != nil {
		return nil, err
	}

	var persons []domain.Person
	for _, record := range result.Records {
		name, _ := record.Get("name")
		born, _ := record.Get("born")
		persons = append(persons, domain.Person{
			Name: name.(string),
			Born: int(born.(int64)),
		})
	}
	return persons, nil
}

func (r *MovieRepository) CreateActedIn(ctx context.Context, actedIn domain.ActedIn) error {
	_, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		`MATCH (p:Person {name: $person_name})
		 MATCH (m:Movie {title: $movie_title})
		 CREATE (p)-[:ACTED_IN {role: $role}]->(m)`,
		map[string]any{
			"person_name": actedIn.PersonName,
			"movie_title": actedIn.MovieTitle,
			"role":        actedIn.Role,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	return err
}

func stringOrEmpty(v any) string {
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}
