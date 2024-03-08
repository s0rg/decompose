package node

type Container struct {
	Cmd    []string          `json:"cmd,omitempty"`
	Env    []string          `json:"env,omitempty"`
	Labels map[string]string `json:"labels"`
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

type Connection struct {
	Port *Port  `json:"port"`
	Src  string `json:"src"`
	Dst  string `json:"dst"`
}

type JSON struct {
	Name       string                   `json:"name"`
	IsExternal bool                     `json:"is_external"`
	Image      *string                  `json:"image,omitempty"`
	Networks   []string                 `json:"networks"`
	Tags       []string                 `json:"tags"`
	Volumes    []*Volume                `json:"volumes"`
	Container  Container                `json:"container"`
	Listen     map[string][]*Port       `json:"listen"`
	Connected  map[string][]*Connection `json:"connected"`
}
