package repository

import (
	"context"
	"fmt"

	"github.com/lwlee2608/learn-neo4j/internal/domain"
	n "github.com/lwlee2608/learn-neo4j/pkg/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type SupplyChainRepository struct {
	client *n.Client
}

func NewSupplyChainRepository(client *n.Client) *SupplyChainRepository {
	return &SupplyChainRepository{client: client}
}

func (r *SupplyChainRepository) CreateCompany(ctx context.Context, company domain.Company) error {
	_, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		"CREATE (c:Company {name: $name, type: $type, founded: $founded, hq: $hq, description: $description})",
		map[string]any{
			"name":        company.Name,
			"type":        company.Type,
			"founded":     company.Founded,
			"hq":          company.HQ,
			"description": company.Description,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	return err
}

func (r *SupplyChainRepository) ListCompanies(ctx context.Context) ([]domain.Company, error) {
	result, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		"MATCH (c:Company) RETURN c.name AS name, c.type AS type, c.founded AS founded, c.hq AS hq, c.description AS description ORDER BY c.name",
		nil,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	if err != nil {
		return nil, err
	}

	var companies []domain.Company
	for _, record := range result.Records {
		name, _ := record.Get("name")
		cType, _ := record.Get("type")
		founded, _ := record.Get("founded")
		hq, _ := record.Get("hq")
		description, _ := record.Get("description")
		companies = append(companies, domain.Company{
			Name:        name.(string),
			Type:        stringOrEmpty(cType),
			Founded:     intOrZero(founded),
			HQ:          stringOrEmpty(hq),
			Description: stringOrEmpty(description),
		})
	}
	return companies, nil
}

func (r *SupplyChainRepository) GetCompany(ctx context.Context, name string) (*domain.CompanyWithRelations, error) {
	result, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		`MATCH (c:Company {name: $name})
		 OPTIONAL MATCH (c)-[:SUPPLIES_EQUIPMENT_TO]->(eq:Company)
		 WITH c, collect(DISTINCT eq.name) AS equipmentClients
		 OPTIONAL MATCH (c)-[:MANUFACTURES_FOR]->(mf:Company)
		 WITH c, equipmentClients, collect(DISTINCT mf.name) AS manufacturingFor
		 OPTIONAL MATCH (c)-[:SUPPLIES_CHIPS_TO]->(sc:Company)
		 WITH c, equipmentClients, manufacturingFor, collect(DISTINCT sc.name) AS chipSuppliedTo
		 OPTIONAL MATCH (c)-[:PROVIDES_CLOUD_FOR]->(cl:Company)
		 WITH c, equipmentClients, manufacturingFor, chipSuppliedTo, collect(DISTINCT cl.name) AS cloudClients
		 OPTIONAL MATCH (c)-[:COMPETES_WITH]-(comp:Company)
		 RETURN c.name AS name, c.type AS type, c.founded AS founded, c.hq AS hq, c.description AS description,
		        equipmentClients, manufacturingFor, chipSuppliedTo, cloudClients,
		        collect(DISTINCT comp.name) AS competitors`,
		map[string]any{"name": name},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	if err != nil {
		return nil, err
	}

	if len(result.Records) == 0 {
		return nil, fmt.Errorf("company not found: %s", name)
	}

	record := result.Records[0]
	cName, _ := record.Get("name")
	cType, _ := record.Get("type")
	founded, _ := record.Get("founded")
	hq, _ := record.Get("hq")
	description, _ := record.Get("description")

	return &domain.CompanyWithRelations{
		Company: domain.Company{
			Name:        cName.(string),
			Type:        stringOrEmpty(cType),
			Founded:     intOrZero(founded),
			HQ:          stringOrEmpty(hq),
			Description: stringOrEmpty(description),
		},
		EquipmentClients: toStringSlice(record, "equipmentClients"),
		ManufacturingFor: toStringSlice(record, "manufacturingFor"),
		ChipSuppliedTo:   toStringSlice(record, "chipSuppliedTo"),
		CloudClients:     toStringSlice(record, "cloudClients"),
		Competitors:      toStringSlice(record, "competitors"),
	}, nil
}

func (r *SupplyChainRepository) CreateSuppliesEquipmentTo(ctx context.Context, rel domain.SuppliesEquipmentTo) error {
	_, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		`MATCH (s:Company {name: $supplier_name})
		 MATCH (r:Company {name: $recipient_name})
		 CREATE (s)-[:SUPPLIES_EQUIPMENT_TO]->(r)`,
		map[string]any{
			"supplier_name":  rel.SupplierName,
			"recipient_name": rel.RecipientName,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	return err
}

func (r *SupplyChainRepository) CreateManufacturesFor(ctx context.Context, rel domain.ManufacturesFor) error {
	_, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		`MATCH (m:Company {name: $manufacturer_name})
		 MATCH (c:Company {name: $client_name})
		 CREATE (m)-[:MANUFACTURES_FOR]->(c)`,
		map[string]any{
			"manufacturer_name": rel.ManufacturerName,
			"client_name":       rel.ClientName,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	return err
}

func (r *SupplyChainRepository) CreateSuppliesChipsTo(ctx context.Context, rel domain.SuppliesChipsTo) error {
	_, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		`MATCH (s:Company {name: $supplier_name})
		 MATCH (c:Company {name: $client_name})
		 CREATE (s)-[:SUPPLIES_CHIPS_TO]->(c)`,
		map[string]any{
			"supplier_name": rel.SupplierName,
			"client_name":   rel.ClientName,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	return err
}

func (r *SupplyChainRepository) CreateCompetesWith(ctx context.Context, rel domain.CompetesWith) error {
	_, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		`MATCH (a:Company {name: $company_name})
		 MATCH (b:Company {name: $competitor_name})
		 CREATE (a)-[:COMPETES_WITH]->(b)`,
		map[string]any{
			"company_name":    rel.CompanyName,
			"competitor_name": rel.CompetitorName,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	return err
}

func (r *SupplyChainRepository) CreateProvidesCloudFor(ctx context.Context, rel domain.ProvidesCloudFor) error {
	_, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		`MATCH (p:Company {name: $provider_name})
		 MATCH (c:Company {name: $client_name})
		 CREATE (p)-[:PROVIDES_CLOUD_FOR]->(c)`,
		map[string]any{
			"provider_name": rel.ProviderName,
			"client_name":   rel.ClientName,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	return err
}

func toStringSlice(record *neo4j.Record, key string) []string {
	raw, _ := record.Get(key)
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	var result []string
	for _, item := range items {
		if s, ok := item.(string); ok && s != "" {
			result = append(result, s)
		}
	}
	return result
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

func intOrZero(v any) int {
	if v == nil {
		return 0
	}
	i, ok := v.(int64)
	if !ok {
		return 0
	}
	return int(i)
}
