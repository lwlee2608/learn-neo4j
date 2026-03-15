package graphexpand

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lwlee2608/learn-neo4j/internal/graphschema"
)

func ValidatePlan(plan *Plan) error {
	allowedRelationshipTypes := graphschema.Default().AllowedRelationshipTypes()

	if plan == nil {
		return errors.New("plan is required")
	}
	if strings.TrimSpace(plan.Keyword) == "" {
		return errors.New("keyword is required")
	}
	if len(plan.Companies) == 0 {
		return errors.New("at least one company is required")
	}

	companies := make(map[string]struct{}, len(plan.Companies))
	for _, company := range plan.Companies {
		name := strings.TrimSpace(company.Name)
		if name == "" {
			return errors.New("company name is required")
		}
		companies[name] = struct{}{}
	}

	for _, rel := range plan.Relationships {
		if _, ok := allowedRelationshipTypes[strings.TrimSpace(rel.Type)]; !ok {
			return fmt.Errorf("relationship type %q is not allowed", rel.Type)
		}
		if strings.TrimSpace(rel.From) == "" || strings.TrimSpace(rel.To) == "" {
			return errors.New("relationship endpoints are required")
		}
		if _, ok := companies[strings.TrimSpace(rel.From)]; !ok {
			return fmt.Errorf("relationship source %q is not listed in companies", rel.From)
		}
		if _, ok := companies[strings.TrimSpace(rel.To)]; !ok {
			return fmt.Errorf("relationship target %q is not listed in companies", rel.To)
		}
	}

	return nil
}
