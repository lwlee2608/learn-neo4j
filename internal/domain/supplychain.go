package domain

type Company struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Founded int    `json:"founded"`
	HQ      string `json:"hq"`
}

type SuppliesEquipmentTo struct {
	SupplierName  string `json:"supplier_name"`
	RecipientName string `json:"recipient_name"`
}

type ManufacturesFor struct {
	ManufacturerName string `json:"manufacturer_name"`
	ClientName       string `json:"client_name"`
}

type SuppliesChipsTo struct {
	SupplierName string `json:"supplier_name"`
	ClientName   string `json:"client_name"`
}

type ProvidesCloudFor struct {
	ProviderName string `json:"provider_name"`
	ClientName   string `json:"client_name"`
}

type CompetesWith struct {
	CompanyName    string `json:"company_name"`
	CompetitorName string `json:"competitor_name"`
}

type CompanyWithRelations struct {
	Company            Company  `json:"company"`
	EquipmentClients   []string `json:"equipment_clients"`
	ManufacturingFor   []string `json:"manufacturing_for"`
	ChipSuppliedTo     []string `json:"chip_supplied_to"`
	CloudClients       []string `json:"cloud_clients"`
	Competitors        []string `json:"competitors"`
}
