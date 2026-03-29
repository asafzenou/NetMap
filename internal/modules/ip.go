package modules

import (
	"context"
	"net"
	"sort"
	"strings"
	"time"

	"recongraph/internal/recon"
)

type IPResult struct {
	IPs []string
}

type IPResolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

type IPModule struct {
	resolver IPResolver
}

func NewIP(resolver IPResolver) *IPModule {
	return &IPModule{resolver: resolver}
}

func (m *IPModule) Name() string { return "ip" }

func (m *IPModule) Run(ctx context.Context, domain string) (recon.ModuleResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	addrs, err := m.resolver.LookupIPAddr(ctx, domain)
	if err != nil {
		return IPResult{}, err
	}

	uniq := map[string]struct{}{}
	for _, a := range addrs {
		ipAddr := a.IP
		if ipAddr == nil {
			continue
		}
		s := strings.TrimSpace(ipAddr.String())
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
	return IPResult{IPs: out}, nil
}
