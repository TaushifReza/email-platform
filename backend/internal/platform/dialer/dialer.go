package dialer

import (
	"context"
	"net"
	"strconv"
	"syscall"
	"time"
)

type Resolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

type Options struct {
	Resolver Resolver
	Timeout  time.Duration
}

type Dialer struct {
	resolver Resolver
	timeout  time.Duration
}

func New(opts Options) *Dialer {
	r := opts.Resolver
	if r == nil {
		r = net.DefaultResolver
	}
	t := opts.Timeout
	if t <= 0 {
		t = 10 * time.Second
	}
	return &Dialer{resolver: r, timeout: t}
}

func (d *Dialer) ValidateTarget(ctx context.Context, host string, port int) error {
	if !portAllowed(port) {
		return ErrNotAllowed
	}
	if ip := net.ParseIP(host); ip != nil {
		if !IsPublicIP(ip) {
			return ErrNotAllowed
		}
		return nil
	}
	addrs, err := d.resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return ErrNotAllowed
	}
	if len(addrs) == 0 {
		return ErrNotAllowed
	}
	for _, a := range addrs {
		if !IsPublicIP(a.IP) {
			return ErrNotAllowed
		}
	}
	return nil
}

func (d *Dialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, ErrNotAllowed
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, ErrNotAllowed
	}
	if !portAllowed(port) {
		return nil, ErrNotAllowed
	}
	if err := d.ValidateTarget(ctx, host, port); err != nil {
		return nil, err
	}
	nd := &net.Dialer{
		Timeout: d.timeout,
		Control: func(_ string, address string, _ syscall.RawConn) error {
			h, p, err := net.SplitHostPort(address)
			if err != nil {
				return ErrNotAllowed
			}
			pp, err := strconv.Atoi(p)
			if err != nil {
				return ErrNotAllowed
			}
			if !portAllowed(pp) {
				return ErrNotAllowed
			}
			ip := net.ParseIP(h)
			if ip == nil {
				addrs, err := d.resolver.LookupIPAddr(ctx, h)
				if err != nil || len(addrs) == 0 {
					return ErrNotAllowed
				}
				for _, a := range addrs {
					if !IsPublicIP(a.IP) {
						return ErrNotAllowed
					}
				}
				return nil
			}
			if !IsPublicIP(ip) {
				return ErrNotAllowed
			}
			return nil
		},
	}
	conn, err := nd.DialContext(ctx, network, addr)
	if err != nil {
		if err.Error() == ErrNotAllowed.Error() {
			return nil, ErrNotAllowed
		}
		return nil, err
	}
	return conn, nil
}
