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
tcp        0      0 127.0.0.1:2333          0.0.0.0:*               LISTEN
tcp        0      0 172.20.4.209:1666       0.0.0.0:*               LISTEN
tcp        0      0 172.20.4.209:48020      172.20.4.198:3306       TIME_WAIT
tcp        0      0 172.20.4.209:1665       172.20.5.76:38512       ESTABLISHED
tcp        1      0 172.20.4.209:43534      172.20.4.129:53         CLOSE_WAIT
tcp        0      0 172.20.4.209:48021      172.20.4.198:3306       ESTABLISHED
tcp6       0      0 :::6501                 :::*                    LISTEN
tcp6       0      0 :::1234                 :::*                    LISTEN
tcp6       0      0 127.0.0.1:6501          127.0.0.1:43706         ESTABLISHED
udp        0      0 127.0.0.11:56688        0.0.0.0:*

some garbage
tcp        0      0 invalid                 172.20.4.198:3306       ESTABLISHED
tcp        0      0 172.20.4.198:3306       invalid                 ESTABLISHED
tcp        0      0 172.20.4.198:bad        172.20.4.198:3306       ESTABLISHED
tcp        0      0 invalid-ip:123          172.20.4.198:3306       ESTABLISHED`)

	res, err := graph.ParseNetstat(b)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) != 5 {
		t.Fail()
	}

	var nlisten, noutbound int

	con := graph.Container{}

	con.SetConnections(res)
	con.ForEachListener(func(_ *graph.Connection) {
		nlisten++
	})
	con.ForEachOutbound(func(_ *graph.Connection) {
		noutbound++
	})

	if nlisten != 3 || noutbound != 1 {
		t.Log(nlisten, noutbound)
		t.Fail()
	}
}

func TestParseNetstatError(t *testing.T) {
	t.Parallel()

	myErr := errors.New("test-err")
	reader := &failReader{Err: myErr}

	_, err := graph.ParseNetstat(reader)
	if err == nil {
		t.Fatal("err == nil")
	}

	if !errors.Is(err, myErr) {
		t.Fail()
	}
}
