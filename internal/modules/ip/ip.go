package ip

import (
	"context"
	"net"
	"sort"
	"strings"
	"time"
)

type Result struct {
	IPs []string
}

type Resolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

type Module struct {
	resolver Resolver
}

func New(resolver Resolver) *Module {
	return &Module{resolver: resolver}
}

func (m *Module) Name() string { return "ip" }

func (m *Module) Run(ctx context.Context, domain string) (any, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	addrs, err := m.resolver.LookupIPAddr(ctx, domain)
	if err != nil {
		return Result{}, err
	}

	uniq := map[string]struct{}{}
	for _, a := range addrs {
		ip := a.IP
		if ip == nil {
			continue
		}
		// Normalize string representation.
		s := strings.TrimSpace(ip.String())
		if s == "" {
			continue
		}
		uniq[s] = struct{}{}
	}

	out := make([]string, 0, len(uniq))
	for s := range uniq {
		out = append(out, s)
	}
	sort.Strings(out)
	return Result{IPs: out}, nil
}

