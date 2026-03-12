package nlquery

import "testing"

func TestValidatePlanAcceptsReadOnlyParameterizedQuery(t *testing.T) {
	plan := &Plan{
		Query:    "MATCH (c:Company)-[:PROVIDES_CLOUD_FOR]->(client:Company) WHERE client.name = $company_name RETURN c.name AS provider ORDER BY provider",
		Params:   map[string]any{"company_name": "OpenAI"},
		ReadOnly: true,
	}

	if err := ValidatePlan(plan, DefaultGraphSchema()); err != nil {
		t.Fatalf("expected plan to be valid, got error: %v", err)
	}
}

func TestValidatePlanRejectsUnsafeKeyword(t *testing.T) {
	plan := &Plan{
		Query:    "MATCH (c:Company) DELETE c RETURN c",
		Params:   map[string]any{},
		ReadOnly: true,
	}

	if err := ValidatePlan(plan, DefaultGraphSchema()); err == nil {
		t.Fatal("expected plan to be rejected")
	}
}

func TestValidatePlanRejectsStringLiteral(t *testing.T) {
	plan := &Plan{
		Query:    "MATCH (c:Company) WHERE c.name = 'OpenAI' RETURN c.name AS name",
		Params:   map[string]any{},
		ReadOnly: true,
	}

	if err := ValidatePlan(plan, DefaultGraphSchema()); err == nil {
		t.Fatal("expected plan to be rejected")
	}
}

func TestValidatePlanRejectsUnknownRelationship(t *testing.T) {
	plan := &Plan{
		Query:    "MATCH (c:Company)-[:DESIGNED]->(d:Company) RETURN c.name AS name",
		Params:   map[string]any{},
		ReadOnly: true,
	}

	if err := ValidatePlan(plan, DefaultGraphSchema()); err == nil {
		t.Fatal("expected plan to be rejected")
	}
}
