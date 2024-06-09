package graph

type NetProto byte

const (
	ALL  NetProto = 0
	TCP  NetProto = 1
	UDP  NetProto = 2
	UNIX NetProto = 3
)

var (
	netKindNames = []string{
		ALL:  "tcp+udp",
		TCP:  "tcp",
		UDP:  "udp",
		UNIX: "unix",
	}

	netKindFlags = []string{
		ALL:  "tux",
		TCP:  "t",
		UDP:  "u",
		UNIX: "x",
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
	case "unix":
		return UNIX, true
	}

	return
}
