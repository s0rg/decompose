package netgraph

type NetProto byte

const (
	ALL NetProto = 0
	TCP NetProto = 1
	UDP NetProto = 2
)

var (
	netKindNames = []string{
		ALL: "tcp+udp",
		TCP: "tcp",
		UDP: "udp",
	}

	netKindFlags = []string{
		ALL: "tu",
		TCP: "t",
		UDP: "u",
	}
)

func (p NetProto) String() string {
	return netKindNames[p]
}

func (p NetProto) Flag() string {
	return netKindFlags[p]
}

func ParseNetProto(val string) (p NetProto, ok bool) {
	switch val {
	case "all":
		return ALL, true
	case "tcp":
		return TCP, true
	case "udp":
		return UDP, true
	}

	return
}
