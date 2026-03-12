package nlquery

import "github.com/lwlee2608/learn-neo4j/internal/graphschema"

type GraphSchema = graphschema.GraphSchema

func DefaultGraphSchema() GraphSchema {
	return graphschema.Default()
}
