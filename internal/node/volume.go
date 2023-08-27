package node

type Volume struct {
	Type string `json:"type"`
	Src  string `json:"src"`
	Dst  string `json:"dst"`
}
