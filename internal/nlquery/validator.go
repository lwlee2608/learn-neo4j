package nlquery

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/lwlee2608/learn-neo4j/internal/graphschema"
)

var (
	forbiddenKeywordPattern = regexp.MustCompile(`(?i)\b(create|merge|delete|detach|set|remove|drop|call|load\s+csv|foreach|apoc|dbms|db\.)\b`)
	labelPattern            = regexp.MustCompile(`\([^\)]*:(\w+)`)
	relationshipPattern     = regexp.MustCompile(`\[:(\w+)`)
	propertyPattern         = regexp.MustCompile(`\b\w+\.(\w+)\b`)
	parameterPattern        = regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)`)
)

func ValidatePlan(plan *Plan, schema graphschema.GraphSchema) error {
	if plan == nil {
		return errors.New("plan is required")
	}

	query := strings.TrimSpace(plan.Query)
	if query == "" {
		return errors.New("query is required")
	}
	if !plan.ReadOnly {
		return errors.New("query plan must be read-only")
	}
	if strings.Contains(query, ";") {
		return errors.New("multiple Cypher statements are not allowed")
	}
	if strings.Contains(query, "//") || strings.Contains(query, "/*") {
		return errors.New("Cypher comments are not allowed")
	}
	if forbiddenKeywordPattern.MatchString(query) {
		return errors.New("query contains forbidden Cypher keywords")
	}

	upperQuery := strings.ToUpper(query)
	if !strings.Contains(upperQuery, "MATCH") {
		return errors.New("query must include MATCH")
	}
	if !strings.Contains(upperQuery, "RETURN") {
		return errors.New("query must include RETURN")
	}
	if strings.Contains(query, "\"") || strings.Contains(query, "'") {
		return errors.New("query must be parameterized and cannot include string literals")
	}

	if err := validateLabels(query, schema); err != nil {
		return err
	}
	if err := validateRelationshipTypes(query, schema); err != nil {
		return err
	}
	if err := validateProperties(query, schema); err != nil {
		return err
	}
	if err := validateParams(query, plan.Params); err != nil {
		return err
	}

	return nil
}

func validateLabels(query string, schema graphschema.GraphSchema) error {
	allowed := schema.AllowedLabels()
	matches := labelPattern.FindAllStringSubmatch(query, -1)
	for _, match := range matches {
		if _, ok := allowed[match[1]]; !ok {
			return fmt.Errorf("label %q is not allowed", match[1])
		}
	}
	return nil
}

func validateRelationshipTypes(query string, schema graphschema.GraphSchema) error {
	allowed := schema.AllowedRelationshipTypes()
	matches := relationshipPattern.FindAllStringSubmatch(query, -1)
	for _, match := range matches {
		if _, ok := allowed[match[1]]; !ok {
			return fmt.Errorf("relationship type %q is not allowed", match[1])
		}
	}
	return nil
}

func validateProperties(query string, schema graphschema.GraphSchema) error {
	allowed := schema.AllowedProperties()
	matches := propertyPattern.FindAllStringSubmatch(query, -1)
	for _, match := range matches {
		prop := match[1]
		if _, ok := allowed[prop]; !ok {
			return fmt.Errorf("property %q is not allowed", prop)
		}
	}
	return nil
}

func validateParams(query string, params map[string]any) error {
	if params == nil {
		params = map[string]any{}
	}

	required := make(map[string]struct{})
	matches := parameterPattern.FindAllStringSubmatch(query, -1)
	for _, match := range matches {
		required[match[1]] = struct{}{}
	}

	for key := range params {
		if _, ok := required[key]; !ok {
			return fmt.Errorf("param %q is not used in query", key)
		}
	}

	for key := range required {
		if _, ok := params[key]; !ok {
			return fmt.Errorf("param %q is required by query", key)
		}
	}

	return nil
}
