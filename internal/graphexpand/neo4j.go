package graphexpand

import (
	"context"
	"fmt"
	"sort"
	"strings"

	n "github.com/lwlee2608/learn-neo4j/pkg/neo4j"
	neo4jdb "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Store interface {
	ListCompanies(ctx context.Context) ([]ExistingCompany, error)
	ApplyPlan(ctx context.Context, plan *Plan) (*ApplyResult, error)
}

type Neo4jStore struct {
	client *n.Client
}

func NewNeo4jStore(client *n.Client) *Neo4jStore {
	return &Neo4jStore{client: client}
}

func (s *Neo4jStore) ListCompanies(ctx context.Context) ([]ExistingCompany, error) {
	result, err := neo4jdb.ExecuteQuery(
		ctx,
		s.client.Driver,
		"MATCH (c:Company) RETURN c.name AS name, c.type AS type ORDER BY c.name",
		nil,
		neo4jdb.EagerResultTransformer,
		neo4jdb.ExecuteQueryWithDatabase("neo4j"),
	)
	if err != nil {
		return nil, fmt.Errorf("list companies: %w", err)
	}

	companies := make([]ExistingCompany, 0, len(result.Records))
	for _, record := range result.Records {
		name, _ := record.Get("name")
		typ, _ := record.Get("type")
		companies = append(companies, ExistingCompany{
			Name: stringValue(name),
			Type: stringValue(typ),
		})
	}

	return companies, nil
}

func (s *Neo4jStore) ApplyPlan(ctx context.Context, plan *Plan) (*ApplyResult, error) {
	if err := ValidatePlan(plan); err != nil {
		return nil, err
	}

	companiesUpserted := make([]string, 0, len(plan.Companies))
	for _, company := range dedupeCompanies(plan.Companies) {
		properties := map[string]any{}
		if strings.TrimSpace(company.Type) != "" {
			properties["type"] = strings.TrimSpace(company.Type)
		}
		if company.Founded > 0 {
			properties["founded"] = company.Founded
		}
		if strings.TrimSpace(company.HQ) != "" {
			properties["hq"] = strings.TrimSpace(company.HQ)
		}

		_, err := neo4jdb.ExecuteQuery(
			ctx,
			s.client.Driver,
			"MERGE (c:Company {name: $name}) SET c += $properties",
			map[string]any{
				"name":       strings.TrimSpace(company.Name),
				"properties": properties,
			},
			neo4jdb.EagerResultTransformer,
			neo4jdb.ExecuteQueryWithDatabase("neo4j"),
		)
		if err != nil {
			return nil, fmt.Errorf("upsert company %q: %w", company.Name, err)
		}
		companiesUpserted = append(companiesUpserted, company.Name)
	}

	relationshipsCreated := make([]string, 0, len(plan.Relationships))
	for _, rel := range plan.Relationships {
		query := fmt.Sprintf(
			"MATCH (a:Company {name: $from}) MATCH (b:Company {name: $to}) MERGE (a)-[:%s]->(b)",
			rel.Type,
		)
		_, err := neo4jdb.ExecuteQuery(
			ctx,
			s.client.Driver,
			query,
			map[string]any{
				"from": rel.From,
				"to":   rel.To,
			},
			neo4jdb.EagerResultTransformer,
			neo4jdb.ExecuteQueryWithDatabase("neo4j"),
		)
		if err != nil {
			return nil, fmt.Errorf("create relationship %q->%q (%s): %w", rel.From, rel.To, rel.Type, err)
		}
		relationshipsCreated = append(relationshipsCreated, fmt.Sprintf("%s -[%s]-> %s", rel.From, rel.Type, rel.To))
	}

	sort.Strings(companiesUpserted)
	sort.Strings(relationshipsCreated)

	return &ApplyResult{
		CompaniesUpserted:    companiesUpserted,
		RelationshipsCreated: relationshipsCreated,
	}, nil
}

func dedupeCompanies(companies []Company) []Company {
	seen := make(map[string]Company, len(companies))
	for _, company := range companies {
		name := strings.TrimSpace(company.Name)
		if name == "" {
			continue
		}
		company.Name = name
		seen[name] = company
	}

	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]Company, 0, len(keys))
	for _, key := range keys {
		result = append(result, seen[key])
	}
	return result
}

func stringValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
