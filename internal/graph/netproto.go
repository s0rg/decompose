package graph

import "strings"

type NetProto int16

const (
	TCP  NetProto = 1 << iota
	UDP  NetProto = 1 << iota
	UNIX NetProto = 1 << iota
	NONE NetProto = 0
	ALL  NetProto = TCP | UDP | UNIX
	pMAX          = 3

	sNONE = "none"
	sTCP  = "tcp"
	sUDP  = "udp"
	sUNIX = "unix"
	sALL  = "all"
	comma = ","
)

func (p NetProto) String() string {
	buf := make([]string, 0, pMAX)

	if p.Has(TCP) {
		buf = append(buf, sTCP)
	}

	if p.Has(UDP) {
		buf = append(buf, sUDP)
	}

	if p.Has(UNIX) {
		buf = append(buf, sUNIX)
	}

	if len(buf) == 0 {
		buf = append(buf, sNONE)
	}

	return strings.Join(buf, comma)
}

func (p *NetProto) Set(mask NetProto) {
	*p |= mask
}

func (p NetProto) Has(mask NetProto) bool {
	return (p & mask) == mask
}

func ParseNetProto(val string) (p NetProto, ok bool) {
	for v := range strings.SplitSeq(val, comma) {
		switch v {
		case sALL:
			p.Set(ALL)
		case sTCP:
			p.Set(TCP)
		case sUDP:
			p.Set(UDP)
		case sUNIX:
			p.Set(UNIX)
		default:
			return
		}
	}

	return p, true
}
