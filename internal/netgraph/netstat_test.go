package netgraph_test

import (
	"bytes"
	"testing"

	"github.com/s0rg/decompose/internal/netgraph"
)

func TestParseNetstat(t *testing.T) {
	t.Parallel()

	b := bytes.NewBufferString(`Active Internet connections (servers and established)
Proto Recv-Q Send-Q Local Address           Foreign Address         State
tcp        0      0 127.0.0.1:2333          0.0.0.0:*               LISTEN
tcp        0      0 172.20.4.209:1666       0.0.0.0:*               LISTEN
tcp        0      0 172.20.4.209:48020      172.20.4.198:3306       TIME_WAIT
tcp        0      0 172.20.4.209:1665       172.20.5.76:38512       ESTABLISHED
tcp        1      0 172.20.4.209:43534      172.20.4.129:53         CLOSE_WAIT
tcp6       0      0 :::6501                 :::*                    LISTEN
tcp6       0      0 :::1234                 :::*                    LISTEN
tcp6       0      0 127.0.0.1:6501          127.0.0.1:43706         ESTABLISHED
udp        0      0 127.0.0.11:56688        0.0.0.0:*

some garbage
tcp        0      0 invalid                 172.20.4.198:3306       ESTABLISHED
tcp        0      0 172.20.4.198:3306       invalid                 ESTABLISHED
tcp        0      0 172.20.4.198:bad        172.20.4.198:3306       ESTABLISHED
tcp        0      0 invalid-ip:123          172.20.4.198:3306       ESTABLISHED`)

	res, err := netgraph.ParseNetstat(b)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) != 4 {
		t.Fail()
	}
}
