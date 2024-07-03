package graph_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/s0rg/decompose/internal/graph"
)

type failReader struct {
	Err error
}

func (fr *failReader) Read(_ []byte) (n int, err error) {
	return 0, fr.Err
}

func TestParseNetstat(t *testing.T) {
	t.Parallel()

	b := bytes.NewBufferString(`Active Internet connections (servers and established)
Proto Recv-Q Send-Q Local Address           Foreign Address         State
tcp        0      0 0.0.0.0:2333            0.0.0.0:*               LISTEN  1/foo
tcp        0      0 172.20.4.209:1666       0.0.0.0:*               LISTEN  2/bar
tcp        0      0 172.20.4.209:48020      172.20.4.198:3306       TIME_WAIT 2/bar
tcp        0      0 172.20.4.209:1665       172.20.5.76:38512       ESTABLISHED 1/foo
tcp        1      0 172.20.4.209:43534      172.20.4.129:53         CLOSE_WAIT 2/bar
tcp        0      0 172.20.4.209:48021      172.20.4.198:3306       ESTABLISHED 1/foo
tcp6       0      0 :::6501                 :::*                    LISTEN 2/bar
tcp6       0      0 :::1234                 :::*                    LISTEN 1/foo
tcp6       0      0 127.0.0.1:6501          127.0.0.1:43706         ESTABLISHED 2/bar
tcp        1      0 172.20.4.209:43634      172.20.4.129:53         ESTABLISHED bar/
tcp        1      0 172.20.4.209:43634      172.20.4.129:53         ESTABLISHED bar
udp        0      0 127.0.0.1:56688         10.10.0.1:54                        11/ntpd
udp        0      0 0.0.0.0:455             0.0.0.0:*                           10/ntpd
bgp        1      1 127.0.0.11:56689        0.0.0.0:*               LISTEN 1/foo
tcp        0      0 invalid                 172.20.4.198:3306       ESTABLISHED -
tcp        0      0 172.20.4.198:3306       invalid                 ESTABLISHED -
tcp        0      0 172.20.4.198:bad        172.20.4.198:3306       ESTABLISHED -
tcp        0      0 invalid-ip:123          172.20.4.198:3306       ESTABLISHED -
Active UNIX domain sockets (servers and established)
Proto RefCnt Flags       Type       State         I-Node   PID/Program name     Path
unix  3      [ ]         STREAM     CONNECTED     38047    1/init               /run/systemd/journal/stdout
unix  3      [ ]         STREAM     CONNECTED     27351    4452/wireplumber
unix  2      [ ACC ]     STREAM     LISTENING     23216    2645/Xorg            @/tmp/.X11-unix/X0
unix  2      [ ]         DGRAM                    39148    4797/xdg-desktop-po

some       garbage
`)

	con := graph.Container{}

	if err := graph.ParseNetstat(b, func(c *graph.Connection) {
		con.AddConnection(c)
	}); err != nil {
		t.Fatal(err)
	}

	if con.ConnectionsCount() != 7 {
		t.Log("total:", con.ConnectionsCount())
		t.Fail()
	}

	var nlisten, noutbound int

	con.IterListeners(func(_ *graph.Connection) {
		nlisten++
	})
	con.IterOutbounds(func(_ *graph.Connection) {
		noutbound++
	})

	if nlisten != 5 || noutbound != 2 {
		t.Log("listen/outbound:", nlisten, noutbound)
		t.Fail()
	}
}

func TestParseNetstatError(t *testing.T) {
	t.Parallel()

	myErr := errors.New("test-err")
	reader := &failReader{Err: myErr}

	err := graph.ParseNetstat(reader, func(*graph.Connection) {})
	if err == nil {
		t.Fatal("err == nil")
	}

	if !errors.Is(err, myErr) {
		t.Fail()
	}
}
