package asn

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

type Result struct {
	ASN string
}

// Provider is an abstraction for IP->ASN resolution.
// MVP ships with a deterministic mock provider; production can add a real provider
// (whois, Team Cymru, paid API, local BGP table, etc.) without touching core orchestration.
type Provider interface {
	LookupASN(ctx context.Context, ip net.IP) (string, error)
	Source() string
}

type Module struct {
	provider Provider
	ips      []string
}

func New(provider Provider, ips []string) *Module {
	return &Module{provider: provider, ips: ips}
}

func (m *Module) Name() string { return "asn" }

func (m *Module) Run(ctx context.Context, _ string) (any, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Select a representative ASN for the domain. For MVP we:
	// - look up each IP
	// - choose the most common ASN
	count := map[string]int{}
	for _, s := range m.ips {
		ip := net.ParseIP(s)
		if ip == nil {
			continue
		}
		asn, err := m.provider.LookupASN(ctx, ip)
		if err != nil {
			continue
		}
		if asn != "" {
			count[asn]++
		}
	}

	bestASN := ""
	bestN := 0
	for asn, n := range count {
		if n > bestN {
			bestASN = asn
			bestN = n
		}
	}

	return Result{ASN: bestASN}, nil
}

// MockProvider produces stable pseudo-ASNs for repeatable MVP behavior.
type MockProvider struct{}

func (p *MockProvider) Source() string { return "mock" }

func (p *MockProvider) LookupASN(_ context.Context, ip net.IP) (string, error) {
	if ip == nil {
		return "", nil
	}
	h := sha256.Sum256([]byte(ip.String()))
	n := binary.BigEndian.Uint32(h[:4])

	// Keep it in a realistic-ish private range (not authoritative).
	asn := 64512 + (n % 1024) // 64512-65535
	return fmt.Sprintf("AS%d", asn), nil
}

