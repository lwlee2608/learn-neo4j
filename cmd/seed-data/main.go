package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	uri := envOrDefault("NEO4J_URI", "bolt://localhost:7687")
	username := envOrDefault("NEO4J_USERNAME", "neo4j")
	password := envOrDefault("NEO4J_PASSWORD", "password")

	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		slog.Error("Failed to create driver", "error", err)
		os.Exit(1)
	}
	defer driver.Close(context.Background())

	ctx := context.Background()

	if err := driver.VerifyConnectivity(ctx); err != nil {
		slog.Error("Failed to connect to Neo4j", "error", err)
		os.Exit(1)
	}
	slog.Info("Connected to Neo4j", "uri", uri)

	// Clear existing data
	run(ctx, driver, "MATCH (n) DETACH DELETE n", nil)
	slog.Info("Cleared existing data")

	// Create indexes
	run(ctx, driver, "CREATE INDEX IF NOT EXISTS FOR (m:Movie) ON (m.title)", nil)
	run(ctx, driver, "CREATE INDEX IF NOT EXISTS FOR (p:Person) ON (p.name)", nil)
	slog.Info("Created indexes")

	// Movies
	movies := []map[string]any{
		{"title": "The Matrix", "released": 1999, "tagline": "Welcome to the Real World"},
		{"title": "The Matrix Reloaded", "released": 2003, "tagline": "Free your mind"},
		{"title": "The Matrix Revolutions", "released": 2003, "tagline": "Everything that has a beginning has an end"},
		{"title": "John Wick", "released": 2014, "tagline": "Don't set him off"},
		{"title": "Speed", "released": 1994, "tagline": "Get ready for rush hour"},
		{"title": "The Devil's Advocate", "released": 1997, "tagline": "Evil has its winning ways"},
		{"title": "A Few Good Men", "released": 1992, "tagline": "You can't handle the truth!"},
		{"title": "Top Gun", "released": 1986, "tagline": "I feel the need, the need for speed"},
		{"title": "Jerry Maguire", "released": 1996, "tagline": "Show me the money"},
		{"title": "Forrest Gump", "released": 1994, "tagline": "Life is like a box of chocolates"},
		{"title": "Cast Away", "released": 2000, "tagline": "At the edge of the world, his journey begins"},
		{"title": "Inception", "released": 2010, "tagline": "Your mind is the scene of the crime"},
		{"title": "The Dark Knight", "released": 2008, "tagline": "Why so serious?"},
		{"title": "Interstellar", "released": 2014, "tagline": "Mankind was born on Earth. It was never meant to die here"},
		{"title": "Fight Club", "released": 1999, "tagline": "The first rule of Fight Club is: you do not talk about Fight Club"},
	}
	for _, m := range movies {
		run(ctx, driver, "CREATE (:Movie {title: $title, released: $released, tagline: $tagline})", m)
	}
	slog.Info("Created movies", "count", len(movies))

	// Persons
	persons := []map[string]any{
		{"name": "Keanu Reeves", "born": 1964},
		{"name": "Laurence Fishburne", "born": 1961},
		{"name": "Carrie-Anne Moss", "born": 1967},
		{"name": "Hugo Weaving", "born": 1960},
		{"name": "Lana Wachowski", "born": 1965},
		{"name": "Lilly Wachowski", "born": 1967},
		{"name": "Tom Cruise", "born": 1962},
		{"name": "Jack Nicholson", "born": 1937},
		{"name": "Tom Hanks", "born": 1956},
		{"name": "Al Pacino", "born": 1940},
		{"name": "Leonardo DiCaprio", "born": 1974},
		{"name": "Christian Bale", "born": 1974},
		{"name": "Heath Ledger", "born": 1979},
		{"name": "Christopher Nolan", "born": 1970},
		{"name": "Brad Pitt", "born": 1963},
		{"name": "Edward Norton", "born": 1969},
		{"name": "David Fincher", "born": 1962},
		{"name": "Robert Zemeckis", "born": 1951},
		{"name": "Chad Stahelski", "born": 1968},
		{"name": "Rob Reiner", "born": 1947},
		{"name": "Tony Scott", "born": 1944},
		{"name": "Cameron Crowe", "born": 1957},
		{"name": "Taylor Hackford", "born": 1944},
		{"name": "Jan de Bont", "born": 1943},
		{"name": "Matthew McConaughey", "born": 1969},
	}
	for _, p := range persons {
		run(ctx, driver, "CREATE (:Person {name: $name, born: $born})", p)
	}
	slog.Info("Created persons", "count", len(persons))

	// ACTED_IN relationships
	actedIn := []map[string]any{
		{"person": "Keanu Reeves", "movie": "The Matrix", "role": "Neo"},
		{"person": "Laurence Fishburne", "movie": "The Matrix", "role": "Morpheus"},
		{"person": "Carrie-Anne Moss", "movie": "The Matrix", "role": "Trinity"},
		{"person": "Hugo Weaving", "movie": "The Matrix", "role": "Agent Smith"},

		{"person": "Keanu Reeves", "movie": "The Matrix Reloaded", "role": "Neo"},
		{"person": "Laurence Fishburne", "movie": "The Matrix Reloaded", "role": "Morpheus"},
		{"person": "Carrie-Anne Moss", "movie": "The Matrix Reloaded", "role": "Trinity"},
		{"person": "Hugo Weaving", "movie": "The Matrix Reloaded", "role": "Agent Smith"},

		{"person": "Keanu Reeves", "movie": "The Matrix Revolutions", "role": "Neo"},
		{"person": "Laurence Fishburne", "movie": "The Matrix Revolutions", "role": "Morpheus"},
		{"person": "Carrie-Anne Moss", "movie": "The Matrix Revolutions", "role": "Trinity"},
		{"person": "Hugo Weaving", "movie": "The Matrix Revolutions", "role": "Agent Smith"},

		{"person": "Keanu Reeves", "movie": "John Wick", "role": "John Wick"},
		{"person": "Keanu Reeves", "movie": "Speed", "role": "Jack Traven"},
		{"person": "Keanu Reeves", "movie": "The Devil's Advocate", "role": "Kevin Lomax"},
		{"person": "Al Pacino", "movie": "The Devil's Advocate", "role": "John Milton"},

		{"person": "Tom Cruise", "movie": "Top Gun", "role": "Maverick"},
		{"person": "Tom Cruise", "movie": "Jerry Maguire", "role": "Jerry Maguire"},
		{"person": "Tom Cruise", "movie": "A Few Good Men", "role": "Lt. Daniel Kaffee"},
		{"person": "Jack Nicholson", "movie": "A Few Good Men", "role": "Col. Nathan R. Jessep"},

		{"person": "Tom Hanks", "movie": "Forrest Gump", "role": "Forrest Gump"},
		{"person": "Tom Hanks", "movie": "Cast Away", "role": "Chuck Noland"},

		{"person": "Leonardo DiCaprio", "movie": "Inception", "role": "Dom Cobb"},
		{"person": "Christian Bale", "movie": "The Dark Knight", "role": "Bruce Wayne"},
		{"person": "Heath Ledger", "movie": "The Dark Knight", "role": "The Joker"},
		{"person": "Leonardo DiCaprio", "movie": "Interstellar", "role": ""},
		{"person": "Matthew McConaughey", "movie": "Interstellar", "role": "Cooper"},

		{"person": "Brad Pitt", "movie": "Fight Club", "role": "Tyler Durden"},
		{"person": "Edward Norton", "movie": "Fight Club", "role": "The Narrator"},
	}
	for _, a := range actedIn {
		run(ctx, driver,
			`MATCH (p:Person {name: $person})
			 MATCH (m:Movie {title: $movie})
			 CREATE (p)-[:ACTED_IN {role: $role}]->(m)`, a)
	}
	slog.Info("Created ACTED_IN relationships", "count", len(actedIn))

	// DIRECTED relationships
	directed := []map[string]any{
		{"person": "Lana Wachowski", "movie": "The Matrix"},
		{"person": "Lilly Wachowski", "movie": "The Matrix"},
		{"person": "Lana Wachowski", "movie": "The Matrix Reloaded"},
		{"person": "Lilly Wachowski", "movie": "The Matrix Reloaded"},
		{"person": "Lana Wachowski", "movie": "The Matrix Revolutions"},
		{"person": "Lilly Wachowski", "movie": "The Matrix Revolutions"},
		{"person": "Chad Stahelski", "movie": "John Wick"},
		{"person": "Jan de Bont", "movie": "Speed"},
		{"person": "Taylor Hackford", "movie": "The Devil's Advocate"},
		{"person": "Rob Reiner", "movie": "A Few Good Men"},
		{"person": "Tony Scott", "movie": "Top Gun"},
		{"person": "Cameron Crowe", "movie": "Jerry Maguire"},
		{"person": "Robert Zemeckis", "movie": "Forrest Gump"},
		{"person": "Robert Zemeckis", "movie": "Cast Away"},
		{"person": "Christopher Nolan", "movie": "Inception"},
		{"person": "Christopher Nolan", "movie": "The Dark Knight"},
		{"person": "Christopher Nolan", "movie": "Interstellar"},
		{"person": "David Fincher", "movie": "Fight Club"},
	}
	for _, d := range directed {
		run(ctx, driver,
			`MATCH (p:Person {name: $person})
			 MATCH (m:Movie {title: $movie})
			 CREATE (p)-[:DIRECTED]->(m)`, d)
	}
	slog.Info("Created DIRECTED relationships", "count", len(directed))

	slog.Info("Seed complete!")
}

func run(ctx context.Context, driver neo4j.DriverWithContext, cypher string, params map[string]any) {
	_, err := neo4j.ExecuteQuery(ctx, driver, cypher, params,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	if err != nil {
		slog.Error("Query failed", "cypher", cypher, "error", err)
		os.Exit(1)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
