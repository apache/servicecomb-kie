package session

//CollectionInfo is struct
type CollectionInfo struct {
	Options Options `json:"options"`
}

//Options is struct
type Options struct {
	Pipeline []map[string]map[string]string `json:"pipeline"`
}
