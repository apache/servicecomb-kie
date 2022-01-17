package model

//ProjectResponse represents the project list
type ProjectResponse struct {
	Total int           `json:"total"`
	Data  []*ProjectDoc `json:"data"`
}
