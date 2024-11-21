package ippool

import (
	"net"
	"net/http"
)

// MakeHTTPClientPool creates a pool of `*http.Client`'s that are evenly
// distributed across the IPs in the given pool.
func MakeHTTPClientPool(pool *Pool, size int) []*http.Client {
	clients, _ := MakePooled(pool, size, func(ip *net.IPAddr) (*http.Client, error) {
		localTCPAddr := net.TCPAddr{
			IP: ip.IP,
		}
		dialer := net.Dialer{
			LocalAddr: &localTCPAddr,
		}

		transport := &http.Transport{
			Dial: dialer.Dial,
		}

		return &http.Client{Transport: transport}, nil
	})

	return clients
}

// MakeHTTPTransportPool creates a pool of `*http.Transport`'s that are evenly
// distributed across the IPs in the given pool.
func MakeHTTPTransportPool(pool *Pool, size int) []*http.Transport {
	transports, _ := MakePooled(pool, size, func(ip *net.IPAddr) (*http.Transport, error) {
		localTCPAddr := net.TCPAddr{
			IP: ip.IP,
		}
		dialer := net.Dialer{
			LocalAddr: &localTCPAddr,
		}

		return &http.Transport{
			Dial: dialer.Dial,
		}, nil
	})

	return transports
}
