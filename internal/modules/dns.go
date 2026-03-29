package modules

import (
	"context"
	"net"
	"sort"
	"strings"
	"time"

	"recongraph/internal/recon"
)

type DNSResult struct {
	NS []string
}

type DNSResolver interface {
	LookupNS(ctx context.Context, name string) ([]*net.NS, error)
}

type DNSModule struct {
	resolver DNSResolver
}

func NewDNS(resolver DNSResolver) *DNSModule {
	return &DNSModule{resolver: resolver}
}

func (m *DNSModule) Name() string { return "dns" }

func (m *DNSModule) Run(ctx context.Context, domain string) (recon.ModuleResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	records, err := m.resolver.LookupNS(ctx, domain)
	if err != nil {
		return DNSResult{}, err
	}

	uniq := map[string]struct{}{}
	for _, r := range records {
		if r == nil {
			continue
		}
		ns := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(r.Host)), ".")
		if ns == "" {
			continue
		}
		uniq[ns] = struct{}{}
	}

	out := make([]string, 0, len(uniq))
	for ns := range uniq {
		out = append(out, ns)
	}
	sort.Strings(out)
	return DNSResult{NS: out}, nil
}
