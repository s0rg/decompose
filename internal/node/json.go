package node

type JSON struct {
	Name       string              `json:"name"`
	Image      *string             `json:"image,omitempty"`
	IsExternal bool                `json:"is_external"`
	Meta       *Meta               `json:"meta,omitempty"`
	Listen     []string            `json:"listen"`
	Networks   []string            `json:"networks"`
	Connected  map[string][]string `json:"connected"`
}
