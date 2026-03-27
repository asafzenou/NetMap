package aggregator

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"recongraph/internal/modules/asn"
	"recongraph/internal/modules/dns"
	"recongraph/internal/modules/ip"
	"recongraph/internal/recon"
	"recongraph/pkg/models"
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
func (a *Aggregator) Collect(ctx context.Context, domain string, asnProvider asn.Provider) (models.ReconOutput, error) {
	if domain == "" {
		return models.ReconOutput{}, errors.New("domain is required")
	}
	if asnProvider == nil {
		asnProvider = &asn.MockProvider{}
	}

	start := time.Now()

	// Factory (MVP): wires the concrete modules behind the ReconModule interface.
	modules := []recon.ReconModule{
		dns.New(a.resolver),
		ip.New(a.resolver),
	}

	resultsCh := make(chan ResultEnvelope, len(modules))
	for _, m := range modules {
		m := m
		go func() {
			val, err := m.Run(ctx, domain)
			resultsCh <- ResultEnvelope{Module: m.Name(), Value: val, Err: err}
		}()
	}

	out := models.ReconOutput{Domain: domain}
	var dnsErr, ipErr error

	for i := 0; i < len(modules); i++ {
		env := <-resultsCh
		switch env.Module {
		case "dns":
			if env.Err != nil {
				dnsErr = env.Err
				continue
			}
			if r, ok := env.Value.(dns.Result); ok {
				out.NS = r.NS
			}
		case "ip":
			if env.Err != nil {
				ipErr = env.Err
				continue
			}
			if r, ok := env.Value.(ip.Result); ok {
				out.IPs = r.IPs
			}
		default:
			// Ignore unknown modules in MVP.
		}
	}

	// ASN depends on IP results, so we run it after collecting IPs.
	asnModule := asn.New(asnProvider, out.IPs)
	asnVal, asnErr := asnModule.Run(ctx, domain)
	if asnErr == nil {
		if r, ok := asnVal.(asn.Result); ok {
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

	// For MVP we tolerate partial failures: return what we could collect,
	// and only error if everything critical failed.
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

