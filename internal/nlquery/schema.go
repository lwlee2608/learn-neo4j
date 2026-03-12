package nlquery

import (
	"fmt"
	"sort"
	"strings"
)

type GraphSchema struct {
	Labels            []string
	RelationshipTypes []string
	Properties        map[string][]string
}

func DefaultGraphSchema() GraphSchema {
	return GraphSchema{
		Labels: []string{"Company"},
		RelationshipTypes: []string{
			"SUPPLIES_EQUIPMENT_TO",
			"MANUFACTURES_FOR",
			"SUPPLIES_CHIPS_TO",
			"PROVIDES_CLOUD_FOR",
			"COMPETES_WITH",
		},
		Properties: map[string][]string{
			"Company": {"name", "type", "founded", "hq"},
		},
	}
}

func (s GraphSchema) Prompt() string {
	var b strings.Builder

	b.WriteString("Graph schema:\n")
	b.WriteString("- Node labels:\n")
	for _, label := range s.sortedLabels() {
		b.WriteString(fmt.Sprintf("  - %s\n", label))
	}

	b.WriteString("- Relationship types:\n")
	for _, rel := range s.sortedRelationshipTypes() {
		b.WriteString(fmt.Sprintf("  - %s\n", rel))
	}

	b.WriteString("- Allowed properties by label:\n")
	for _, label := range s.sortedPropertyLabels() {
		b.WriteString(fmt.Sprintf("  - %s: %s\n", label, strings.Join(s.Properties[label], ", ")))
	}

	b.WriteString("- Domain notes:\n")
	b.WriteString("  - Companies represent chip designers, manufacturers, equipment suppliers, AI labs, and cloud providers.\n")
	b.WriteString("  - The graph currently models company-to-company relationships only.\n")
	b.WriteString("  - When querying relationships for a specific entity, use undirected patterns like (a)-[r]-(b) to capture both incoming and outgoing relationships in a single query.\n")
	b.WriteString("  - Use only the labels, relationships, and properties listed here.\n")

	return b.String()
}

func (s GraphSchema) allowedLabels() map[string]struct{} {
	allowed := make(map[string]struct{}, len(s.Labels))
	for _, label := range s.Labels {
		allowed[label] = struct{}{}
	}
	return allowed
}

func (s GraphSchema) allowedRelationshipTypes() map[string]struct{} {
	allowed := make(map[string]struct{}, len(s.RelationshipTypes))
	for _, rel := range s.RelationshipTypes {
		allowed[rel] = struct{}{}
	}
	return allowed
}

func (s GraphSchema) allowedProperties() map[string]struct{} {
	allowed := make(map[string]struct{})
	for _, props := range s.Properties {
		for _, prop := range props {
			allowed[prop] = struct{}{}
		}
	}
	return allowed
}

func (s GraphSchema) sortedLabels() []string {
	labels := append([]string(nil), s.Labels...)
	sort.Strings(labels)
	return labels
}

func (s GraphSchema) sortedRelationshipTypes() []string {
	relationships := append([]string(nil), s.RelationshipTypes...)
	sort.Strings(relationships)
	return relationships
}

func (s GraphSchema) sortedPropertyLabels() []string {
	labels := make([]string, 0, len(s.Properties))
	for label := range s.Properties {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	return labels
}
