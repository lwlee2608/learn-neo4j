package nlquery

type Plan struct {
	Query       string         `json:"query"`
	Params      map[string]any `json:"params"`
	Intent      string         `json:"intent"`
	Explanation string         `json:"explanation"`
	ReadOnly    bool           `json:"read_only"`
}

type QueryResult struct {
	Records []map[string]any `json:"records"`
	Count   int              `json:"count"`
}

type Answer struct {
	Plan          *Plan        `json:"plan,omitempty"`
	Result        *QueryResult `json:"result,omitempty"`
	FinalResponse string       `json:"final_response"`
}
