package node

type PortMatcher interface {
	HasAny(...string) bool
	Has(...string) bool
}

type View struct {
	Listen     PortMatcher
	Name       string
	Image      string
	Cmd        string
	Args       []string
	Tags       []string
	IsExternal bool
}
