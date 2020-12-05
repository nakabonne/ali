package attacker

import (
	"context"
	"net"
	"sync/atomic"
	"time"
)

type resolver struct {
	addrs []string
	idx   uint64
}

func NewResolver(addrs []string) *net.Resolver {
	r := &resolver{addrs: addrs}

	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(2000),
			}
			return d.DialContext(ctx, network, r.address())
		},
	}
}

func (r *resolver) address() string {
	return r.addrs[atomic.AddUint64(&r.idx, 1)%uint64(len(r.addrs))]
}
