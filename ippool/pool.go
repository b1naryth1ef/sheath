package ippool

import "net"

type Pool struct {
	ips []*net.IPAddr
}

func (p *Pool) Size() int {
	return len(p.ips)
}

func (p *Pool) N(n int) *net.IPAddr {
	if len(p.ips) == 0 {
		return nil
	}

	idx := n % len(p.ips)
	return p.ips[idx]
}

func (p *Pool) AddString(ip string) error {
	addr, err := net.ResolveIPAddr("ip", ip)
	if err != nil {
		return err
	}
	p.Add(addr)
	return nil
}

func (p *Pool) Add(ip *net.IPAddr) {
	p.ips = append(p.ips, ip)
}

func NewPool() *Pool {
	return &Pool{
		ips: make([]*net.IPAddr, 0),
	}
}
