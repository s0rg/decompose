package node

type Process struct {
	Cmd []string `json:"cmd"`
	Env []string `json:"env"`
}

type Volume struct {
	Type string `json:"type"`
	Src  string `json:"src"`
	Dst  string `json:"dst"`
}

type Meta struct {
	Info string   `json:"info"`
	Docs string   `json:"docs"`
	Repo string   `json:"repo"`
	Tags []string `json:"tags"`
}

type JSON struct {
	Name       string              `json:"name"`
	IsExternal bool                `json:"is_external"`
	Image      *string             `json:"image,omitempty"`
	Process    *Process            `json:"process,omitempty"`
	Listen     []string            `json:"listen"`
	Networks   []string            `json:"networks"`
	Tags       []string            `json:"tags"`
	Volumes    []*Volume           `json:"volumes"`
	Connected  map[string][]string `json:"connected"`
}
