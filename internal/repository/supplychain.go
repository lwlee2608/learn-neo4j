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
		"CREATE (c:Company {name: $name, type: $type, founded: $founded, hq: $hq})",
		map[string]any{
			"name":    company.Name,
			"type":    company.Type,
			"founded": company.Founded,
			"hq":      company.HQ,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	return err
}

func (r *SupplyChainRepository) ListCompanies(ctx context.Context) ([]domain.Company, error) {
	result, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		"MATCH (c:Company) RETURN c.name AS name, c.type AS type, c.founded AS founded, c.hq AS hq ORDER BY c.name",
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
		companies = append(companies, domain.Company{
			Name:    name.(string),
			Type:    stringOrEmpty(cType),
			Founded: intOrZero(founded),
			HQ:      stringOrEmpty(hq),
		})
	}
	return companies, nil
}

func (r *SupplyChainRepository) GetCompany(ctx context.Context, name string) (*domain.CompanyWithRelations, error) {
	result, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		`MATCH (c:Company {name: $name})
		 OPTIONAL MATCH (c)-[:DESIGNED]->(ch1:Chip)
		 WITH c, collect(DISTINCT ch1.name) AS chipsDesigned
		 OPTIONAL MATCH (c)-[:MANUFACTURES]->(ch2:Chip)
		 WITH c, chipsDesigned, collect(DISTINCT ch2.name) AS chipsManufactured
		 OPTIONAL MATCH (c)-[:SUPPLIES_EQUIPMENT_TO]->(r:Company)
		 WITH c, chipsDesigned, chipsManufactured, collect(DISTINCT r.name) AS equipmentClients
		 OPTIONAL MATCH (c)-[:PROVIDES_CLOUD_FOR]->(cl:Company)
		 WITH c, chipsDesigned, chipsManufactured, equipmentClients, collect(DISTINCT cl.name) AS cloudClients
		 OPTIONAL MATCH (c)-[:USES]->(ch3:Chip)
		 RETURN c.name AS name, c.type AS type, c.founded AS founded, c.hq AS hq,
		        chipsDesigned, chipsManufactured, equipmentClients, cloudClients,
		        collect(DISTINCT ch3.name) AS chipsUsed`,
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

	return &domain.CompanyWithRelations{
		Company: domain.Company{
			Name:    cName.(string),
			Type:    stringOrEmpty(cType),
			Founded: intOrZero(founded),
			HQ:      stringOrEmpty(hq),
		},
		ChipsDesigned:     toStringSlice(record, "chipsDesigned"),
		ChipsManufactured: toStringSlice(record, "chipsManufactured"),
		EquipmentClients:  toStringSlice(record, "equipmentClients"),
		CloudClients:      toStringSlice(record, "cloudClients"),
		ChipsUsed:         toStringSlice(record, "chipsUsed"),
	}, nil
}

func (r *SupplyChainRepository) CreateChip(ctx context.Context, chip domain.Chip) error {
	_, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		"CREATE (ch:Chip {name: $name, architecture: $architecture, year: $year, transistor_nm: $transistor_nm})",
		map[string]any{
			"name":          chip.Name,
			"architecture":  chip.Architecture,
			"year":          chip.Year,
			"transistor_nm": chip.TransistorNm,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	return err
}

func (r *SupplyChainRepository) ListChips(ctx context.Context) ([]domain.Chip, error) {
	result, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		"MATCH (ch:Chip) RETURN ch.name AS name, ch.architecture AS architecture, ch.year AS year, ch.transistor_nm AS transistor_nm ORDER BY ch.year",
		nil,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	if err != nil {
		return nil, err
	}

	var chips []domain.Chip
	for _, record := range result.Records {
		name, _ := record.Get("name")
		arch, _ := record.Get("architecture")
		year, _ := record.Get("year")
		nm, _ := record.Get("transistor_nm")
		chips = append(chips, domain.Chip{
			Name:         name.(string),
			Architecture: stringOrEmpty(arch),
			Year:         intOrZero(year),
			TransistorNm: intOrZero(nm),
		})
	}
	return chips, nil
}

func (r *SupplyChainRepository) GetChip(ctx context.Context, name string) (*domain.ChipWithRelations, error) {
	result, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		`MATCH (ch:Chip {name: $name})
		 OPTIONAL MATCH (d:Company)-[:DESIGNED]->(ch)
		 WITH ch, collect(DISTINCT d.name) AS designers
		 OPTIONAL MATCH (m:Company)-[:MANUFACTURES]->(ch)
		 WITH ch, designers, collect(DISTINCT m.name) AS manufacturers
		 OPTIONAL MATCH (u:Company)-[:USES]->(ch)
		 RETURN ch.name AS name, ch.architecture AS architecture, ch.year AS year, ch.transistor_nm AS transistor_nm,
		        designers, manufacturers, collect(DISTINCT u.name) AS users`,
		map[string]any{"name": name},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	if err != nil {
		return nil, err
	}

	if len(result.Records) == 0 {
		return nil, fmt.Errorf("chip not found: %s", name)
	}

	record := result.Records[0]
	cName, _ := record.Get("name")
	arch, _ := record.Get("architecture")
	year, _ := record.Get("year")
	nm, _ := record.Get("transistor_nm")

	return &domain.ChipWithRelations{
		Chip: domain.Chip{
			Name:         cName.(string),
			Architecture: stringOrEmpty(arch),
			Year:         intOrZero(year),
			TransistorNm: intOrZero(nm),
		},
		Designers:     toStringSlice(record, "designers"),
		Manufacturers: toStringSlice(record, "manufacturers"),
		Users:         toStringSlice(record, "users"),
	}, nil
}

func (r *SupplyChainRepository) CreateDesigned(ctx context.Context, rel domain.Designed) error {
	_, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		`MATCH (c:Company {name: $company_name})
		 MATCH (ch:Chip {name: $chip_name})
		 CREATE (c)-[:DESIGNED]->(ch)`,
		map[string]any{
			"company_name": rel.CompanyName,
			"chip_name":    rel.ChipName,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	return err
}

func (r *SupplyChainRepository) CreateManufactures(ctx context.Context, rel domain.Manufactures) error {
	_, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		`MATCH (c:Company {name: $company_name})
		 MATCH (ch:Chip {name: $chip_name})
		 CREATE (c)-[:MANUFACTURES]->(ch)`,
		map[string]any{
			"company_name": rel.CompanyName,
			"chip_name":    rel.ChipName,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	return err
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

func (r *SupplyChainRepository) CreateUses(ctx context.Context, rel domain.Uses) error {
	_, err := neo4j.ExecuteQuery(ctx, r.client.Driver,
		`MATCH (c:Company {name: $company_name})
		 MATCH (ch:Chip {name: $chip_name})
		 CREATE (c)-[:USES]->(ch)`,
		map[string]any{
			"company_name": rel.CompanyName,
			"chip_name":    rel.ChipName,
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
