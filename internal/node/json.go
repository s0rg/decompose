package node

type JSON struct {
	Name       string              `json:"name"`
	IsExternal bool                `json:"is_external"`
	Image      *string             `json:"image,omitempty"`
	Meta       *Meta               `json:"meta,omitempty"`
	Process    *Process            `json:"process,omitempty"`
	Listen     []string            `json:"listen"`
	Networks   []string            `json:"networks"`
	Volumes    []*Volume           `json:"volumes"`
	Connected  map[string][]string `json:"connected"`
}
