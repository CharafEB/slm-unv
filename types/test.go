package types

type Input struct {
	Name string `json:"name" jsonschema:"the name of the person to greet"`
}

type Output struct {
	Greeting string `json:"greeting" jsonschema:"the greeting to tell to the user"`
}

type DBSearchInput struct {
	Term    string   `json:"term" jsonschema_description:"The search term."`
	Columns []string `json:"columns,omitempty" jsonschema_description:"The columns to search in. Defaults to all columns."`
}


type DBSearchOutput struct {
	Found   bool     `json:"found"`
	Results []string `json:"results,omitempty"`
	Source  string   `json:"source"`
}