package main

import (
	"net"
	"strconv"
	"strings"

	"github.com/aerospike/aerospike-client-go"
)

const aeroDefaultPort = 3000

type hostsListValue []*aerospike.Host

func newHostsListValue(value string, p *[]*aerospike.Host) *hostsListValue {
	l := (*hostsListValue)(p)
	l.Set(value)
	return l
}

func (l hostsListValue) String() string {
	var buf strings.Builder
	for n, host := range l {
		if n > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(host.String())
	}
	return buf.String()
}

func (l *hostsListValue) Set(v string) (err error) {
	h := strings.Split(v, ",")
	list := make([]*aerospike.Host, len(h))
	for i, hostport := range h {
		var (
			host string
			port int
		)
		if strings.ContainsRune(hostport, ':') {
			var rawPort string
			host, rawPort, err = net.SplitHostPort(hostport)
			if err != nil {
				return err
			}
			port, err = strconv.Atoi(rawPort)
			if err != nil {
				return err
			}
		} else {
			host = hostport
			port = aeroDefaultPort
		}
		list[i] = aerospike.NewHost(host, port)
	}
	*l = list
	return nil
}
