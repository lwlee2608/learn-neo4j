package graphexpand

type Company struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Founded int    `json:"founded,omitempty"`
	HQ      string `json:"hq,omitempty"`
}

type Relationship struct {
	Type      string `json:"type"`
	From      string `json:"from"`
	To        string `json:"to"`
	Rationale string `json:"rationale,omitempty"`
}

type Plan struct {
	Keyword       string         `json:"keyword"`
	Summary       string         `json:"summary"`
	Companies     []Company      `json:"companies"`
	Relationships []Relationship `json:"relationships"`
	Sources       []string       `json:"sources,omitempty"`
}

type ApplyResult struct {
	CompaniesUpserted    []string `json:"companies_upserted"`
	RelationshipsCreated []string `json:"relationships_created"`
}

type Answer struct {
	Plan          *Plan        `json:"plan,omitempty"`
	ApplyResult   *ApplyResult `json:"apply_result,omitempty"`
	FinalResponse string       `json:"final_response"`
}

type ExistingCompany struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
