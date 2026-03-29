package aggregator

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"recongraph/recon/engine/modules"
	"recongraph/recon/engine/recon"
	"recongraph/recon/pkg/models"
)

type ResultEnvelope struct {
	Module string
	Value  recon.ModuleResult
	Err    error
}

type Aggregator struct {
	resolver *net.Resolver
}

func New(resolver *net.Resolver) *Aggregator {
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	return &Aggregator{resolver: resolver}
}

// Collect performs recon for a single domain and returns the JSON contract.
func (a *Aggregator) Collect(ctx context.Context, domain string, asnProvider modules.ASNProvider) (models.ReconOutput, error) {
	if domain == "" {
		return models.ReconOutput{}, errors.New("domain is required")
	}
	if asnProvider == nil {
		asnProvider = &modules.MockASNProvider{}
	}

	start := time.Now()

	mods := []recon.ReconModule{
		modules.NewDNS(a.resolver),
		modules.NewIP(a.resolver),
	}

	resultsCh := make(chan ResultEnvelope, len(mods))
	for _, m := range mods {
		m := m
		go func() {
			val, err := m.Run(ctx, domain)
			resultsCh <- ResultEnvelope{Module: m.Name(), Value: val, Err: err}
		}()
	}

	out := models.ReconOutput{Domain: domain}
	var dnsErr, ipErr error

	for i := 0; i < len(mods); i++ {
		env := <-resultsCh
		switch env.Module {
		case "dns":
			if env.Err != nil {
				dnsErr = env.Err
				continue
			}
			if r, ok := env.Value.(modules.DNSResult); ok {
				out.NS = r.NS
			}
		case "ip":
			if env.Err != nil {
				ipErr = env.Err
				continue
			}
			if r, ok := env.Value.(modules.IPResult); ok {
				out.IPs = r.IPs
			}
		default:
			// Ignore unknown modules in MVP.
		}
	}

	asnModule := modules.NewASN(asnProvider, out.IPs)
	asnVal, asnErr := asnModule.Run(ctx, domain)
	if asnErr == nil {
		if r, ok := asnVal.(modules.ASNResult); ok {
			out.ASN = r.ASN
		}
	}

	out.Metadata.CollectedAt = time.Now().UTC()
	out.Metadata.DurationMs = time.Since(start).Milliseconds()
	out.Metadata.Sources = []string{
		"dns:net.Resolver",
		"ip:net.Resolver",
		fmt.Sprintf("asn:%s", asnProvider.Source()),
	}

	if len(out.NS) == 0 && len(out.IPs) == 0 {
		if ipErr != nil {
			return out, fmt.Errorf("recon failed (ip): %w", ipErr)
		}
		if dnsErr != nil {
			return out, fmt.Errorf("recon failed (dns): %w", dnsErr)
		}
		return out, errors.New("recon failed: no results")
	}

	return out, nil
}
