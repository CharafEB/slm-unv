package types

// Author struct
type AuthorInput struct {
	Name string `json:"name" jsonschema_description:"the auther name to search on.`
}

type AuthorOutput struct {
	Found   bool    `json:"found"`
	Results []string `json:"results,omitempty"`
}

// Article struct
type ArticleInput struct {
	Title string `json:"name" jsonschema_description:"the auther name to search on.`
}

type ArticleOutput struct {
	Found   bool     `json:"found"`
	Results []string `json:"results,omitempty"`
}
