package node

type Process struct {
	Cmd []string `json:"cmd"`
	Env []string `json:"env"`
}
