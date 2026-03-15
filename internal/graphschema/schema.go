package graphschema

import (
	"bytes"
	_ "embed"
	"sort"
	"strings"
	"text/template"
)

type GraphSchema struct {
	Labels            []string
	RelationshipTypes []string
	Properties        map[string][]string
}

func Default() GraphSchema {
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
			"Company": {"name", "type", "founded", "hq", "description"},
		},
	}
}

//go:embed templates/schema_prompt.tmpl
var schemaPromptRaw string

var schemaPromptTmpl = template.Must(template.New("schema").Funcs(template.FuncMap{
	"join": strings.Join,
}).Parse(schemaPromptRaw))

type labelProps struct {
	Label string
	Props []string
}

func (s GraphSchema) Prompt() string {
	propLabels := make([]labelProps, 0, len(s.Properties))
	for _, label := range s.sortedPropertyLabels() {
		propLabels = append(propLabels, labelProps{Label: label, Props: s.Properties[label]})
	}

	data := struct {
		Labels            []string
		RelationshipTypes []string
		PropertyLabels    []labelProps
	}{
		Labels:            s.sortedLabels(),
		RelationshipTypes: s.sortedRelationshipTypes(),
		PropertyLabels:    propLabels,
	}

	var buf bytes.Buffer
	schemaPromptTmpl.Execute(&buf, data)
	return buf.String()
}

func (s GraphSchema) AllowedLabels() map[string]struct{} {
	allowed := make(map[string]struct{}, len(s.Labels))
	for _, label := range s.Labels {
		allowed[label] = struct{}{}
	}
	return allowed
}

func (s GraphSchema) AllowedRelationshipTypes() map[string]struct{} {
	allowed := make(map[string]struct{}, len(s.RelationshipTypes))
	for _, rel := range s.RelationshipTypes {
		allowed[rel] = struct{}{}
	}
	return allowed
}

func (s GraphSchema) AllowedProperties() map[string]struct{} {
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
