package ippool

import (
	"net"
	"net/http"
)

func MakeHTTPClientPool(pool *Pool, size int) []*http.Client {
	if size <= 0 {
		size = pool.Size()
	}

	clients := make([]*http.Client, 0, size)
	for n := range size {
		ip := pool.N(n)
		localTCPAddr := net.TCPAddr{
			IP: ip.IP,
		}
		dialer := net.Dialer{
			LocalAddr: &localTCPAddr,
		}

		transport := &http.Transport{
			Dial: dialer.Dial,
		}

		clients = append(clients, &http.Client{Transport: transport})
	}

	return clients
}
