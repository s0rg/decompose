package node

type JSON struct {
	Name       string              `json:"name"`
	Image      *string             `json:"image,omitempty"`
	IsExternal bool                `json:"is_external"`
	Listen     []string            `json:"listen"`
	Connected  map[string][]string `json:"connected"`
}
