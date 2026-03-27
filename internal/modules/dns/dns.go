package dns

import (
	"context"
	"net"
	"sort"
	"strings"
	"time"
)

type Result struct {
	NS []string
}

type Resolver interface {
	LookupNS(ctx context.Context, name string) ([]*net.NS, error)
}

type Module struct {
	resolver Resolver
}

func New(resolver Resolver) *Module {
	return &Module{resolver: resolver}
}

func (m *Module) Name() string { return "dns" }

func (m *Module) Run(ctx context.Context, domain string) (any, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	records, err := m.resolver.LookupNS(ctx, domain)
	if err != nil {
		return Result{}, err
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
	return Result{NS: out}, nil
}

