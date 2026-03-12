package graphexpand

import "testing"

func TestValidatePlan(t *testing.T) {
	plan := &Plan{
		Keyword: "CoreWeave",
		Summary: "Adds CoreWeave and links it to OpenAI.",
		Companies: []Company{
			{Name: "CoreWeave", Type: "cloud_provider"},
			{Name: "OpenAI", Type: "ai_lab"},
		},
		Relationships: []Relationship{
			{Type: "PROVIDES_CLOUD_FOR", From: "CoreWeave", To: "OpenAI"},
		},
	}

	if err := ValidatePlan(plan); err != nil {
		t.Fatalf("expected plan to be valid, got %v", err)
	}
}

func TestValidatePlanRejectsUnknownRelationship(t *testing.T) {
	plan := &Plan{
		Keyword:   "CoreWeave",
		Companies: []Company{{Name: "CoreWeave"}, {Name: "OpenAI"}},
		Relationships: []Relationship{
			{Type: "USES", From: "CoreWeave", To: "OpenAI"},
		},
	}

	if err := ValidatePlan(plan); err == nil {
		t.Fatal("expected plan to be rejected")
	}
}
