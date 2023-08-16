package netgraph

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"slices"
	"strconv"
	"strings"
)

const (
	stateListen      = "LISTEN"
	stateEstablished = "ESTABLISHED"
)

func ParseNetstat(r io.Reader) (rv []*Connection, err error) {
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanLines)

	const (
		nSkipHead = 2
		nConnsNum = 64
	)

	var (
		conn *Connection
		ok   bool
	)

	rv = make([]*Connection, 0, nConnsNum)

	for i := 0; s.Scan(); i++ {
		if i < nSkipHead {
			continue
		}

		if conn, ok = parseConnection(s.Text()); !ok {
			continue
		}

		rv = append(rv, conn)
	}

	if err = s.Err(); err != nil {
		return nil, fmt.Errorf("scan: %w", err)
	}

	return slices.Clip(rv), nil
}

func parseConnection(s string) (conn *Connection, ok bool) {
	parts := strings.Fields(s)
	if len(parts) < 1 {
		return nil, false
	}

	conn = &Connection{}

	if conn.Kind, ok = parseKind(parts[0], len(parts)); !ok {
		return nil, false
	}

	if conn.LocalIP, conn.LocalPort, ok = splitIP(parts[3]); !ok {
		return nil, false
	}

	if conn.LocalIP.IsLoopback() {
		return nil, false
	}

	if conn.RemoteIP, conn.RemotePort, ok = splitIP(parts[4]); !ok {
		return nil, false
	}

	if conn.Kind == TCP {
		switch parts[5] {
		case stateListen, stateEstablished:
		default: // skip all other states
			return nil, false
		}
	}

	return conn, true
}

func parseKind(kind string, fieldsNum int) (k NetProto, ok bool) {
	const (
		nPartsUDP = 5
		nPartsTCP = 6
	)

	switch {
	case strings.HasPrefix(kind, TCP.String()) && fieldsNum == nPartsTCP:
		return TCP, true
	case strings.HasPrefix(kind, UDP.String()) && fieldsNum == nPartsUDP:
		return UDP, true
	default: // unknown protocol or invalid fields count
	}

	return
}

func splitIP(v string) (ip net.IP, port uint16, ok bool) {
	idx := strings.LastIndexByte(v, ':')
	if idx < 0 {
		return
	}

	addr, sport := v[:idx], v[idx+1:]

	if ip = net.ParseIP(addr); ip == nil {
		return
	}

	if sport != "*" {
		uval, err := strconv.ParseUint(sport, 10, 16)
		if err != nil {
			return
		}

		port = uint16(uval)
	}

	return ip, port, true
}
