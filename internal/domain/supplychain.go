package domain

type Company struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Founded int    `json:"founded"`
	HQ      string `json:"hq"`
}

type Chip struct {
	Name         string `json:"name"`
	Architecture string `json:"architecture"`
	Year         int    `json:"year"`
	TransistorNm int   `json:"transistor_nm"`
}

type Designed struct {
	CompanyName string `json:"company_name"`
	ChipName    string `json:"chip_name"`
}

type Manufactures struct {
	CompanyName string `json:"company_name"`
	ChipName    string `json:"chip_name"`
}

type SuppliesEquipmentTo struct {
	SupplierName  string `json:"supplier_name"`
	RecipientName string `json:"recipient_name"`
}

type ProvidesCloudFor struct {
	ProviderName string `json:"provider_name"`
	ClientName   string `json:"client_name"`
}

type Uses struct {
	CompanyName string `json:"company_name"`
	ChipName    string `json:"chip_name"`
}

type CompanyWithRelations struct {
	Company           Company  `json:"company"`
	ChipsDesigned     []string `json:"chips_designed"`
	ChipsManufactured []string `json:"chips_manufactured"`
	EquipmentClients  []string `json:"equipment_clients"`
	CloudClients      []string `json:"cloud_clients"`
	ChipsUsed         []string `json:"chips_used"`
}

type ChipWithRelations struct {
	Chip          Chip     `json:"chip"`
	Designers     []string `json:"designers"`
	Manufacturers []string `json:"manufacturers"`
	Users         []string `json:"users"`
}
