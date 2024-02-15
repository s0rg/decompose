package graph

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

const (
	stateListen      = "LISTEN"
	stateEstablished = "ESTABLISHED"
)

func ParseNetstat(r io.Reader, cb func(*Connection)) (err error) {
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanLines)

	const nSkipHead = 2

	var (
		conn *Connection
		ok   bool
	)

	for i := 0; s.Scan(); i++ {
		if i < nSkipHead {
			continue
		}

		if conn, ok = parseConnection(s.Text()); !ok {
			continue
		}

		cb(conn)
	}

	if err = s.Err(); err != nil {
		return fmt.Errorf("scan: %w", err)
	}

	return nil
}

func parseConnection(s string) (conn *Connection, ok bool) {
	const minFields = 6

	parts := strings.Fields(s)
	if len(parts) < minFields {
		return nil, false
	}

	conn = &Connection{}

	if conn.Proto, ok = parseKind(parts[0], len(parts)); !ok {
		return nil, false
	}

	if conn.LocalIP, conn.LocalPort, ok = splitIP(parts[3]); !ok {
		return nil, false
	}

	if conn.RemoteIP, conn.RemotePort, ok = splitIP(parts[4]); !ok {
		return nil, false
	}

	var nProcField = 5

	if conn.Proto == TCP {
		nProcField = 6

		switch parts[5] {
		case stateListen, stateEstablished:
		default: // skip all other states
			return nil, false
		}
	}

	if conn.Process, ok = splitName(parts[nProcField]); !ok {
		return nil, false
	}

	return conn, true
}

func parseKind(kind string, fieldsNum int) (k NetProto, ok bool) {
	const (
		nPartsUDP = 6
		nPartsTCP = 7
	)

	switch {
	case strings.HasPrefix(kind, TCP.String()) && fieldsNum >= nPartsTCP:
		return TCP, true
	case strings.HasPrefix(kind, UDP.String()) && fieldsNum >= nPartsUDP:
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

func splitName(v string) (name string, ok bool) {
	const pidFields = 2

	if !strings.ContainsRune(v, '/') {
		return
	}

	parts := strings.SplitN(v, "/", pidFields)
	if len(parts) != pidFields {
		return
	}

	fields := strings.Fields(parts[1])

	name = fields[0]

	return name, name != ""
}
