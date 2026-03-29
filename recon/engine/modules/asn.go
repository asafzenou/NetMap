package modules

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"recongraph/recon/engine/recon"
)

type ASNResult struct {
	ASN string
}

// ASNProvider resolves IP addresses to ASN strings.
type ASNProvider interface {
	LookupASN(ctx context.Context, ip net.IP) (string, error)
	Source() string
}

type ASNModule struct {
	provider ASNProvider
	ips      []string
}

func NewASN(provider ASNProvider, ips []string) *ASNModule {
	return &ASNModule{provider: provider, ips: ips}
}

func (m *ASNModule) Name() string { return "asn" }

func (m *ASNModule) Run(ctx context.Context, _ string) (recon.ModuleResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	count := map[string]int{}
	for _, s := range m.ips {
		ipAddr := net.ParseIP(s)
		if ipAddr == nil {
			continue
		}
		asnStr, err := m.provider.LookupASN(ctx, ipAddr)
		if err != nil {
			continue
		}
		if asnStr != "" {
			count[asnStr]++
		}
	}

	bestASN := ""
	bestN := 0
	for asnStr, n := range count {
		if n > bestN {
			bestASN = asnStr
			bestN = n
		}
	}

	return ASNResult{ASN: bestASN}, nil
}

// MockASNProvider produces stable pseudo-ASNs for repeatable MVP behavior.
type MockASNProvider struct{}

func (p *MockASNProvider) Source() string { return "mock" }

func (p *MockASNProvider) LookupASN(_ context.Context, ip net.IP) (string, error) {
	if ip == nil {
		return "", nil
	}
	h := sha256.Sum256([]byte(ip.String()))
	n := binary.BigEndian.Uint32(h[:4])
	asnNum := 64512 + (n % 1024)
	return fmt.Sprintf("AS%d", asnNum), nil
}
