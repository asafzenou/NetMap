package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	aggregator "recongraph/recon/engine"
	"recongraph/recon/engine/modules"
)

// parseFlags parses and validates command-line arguments.
func parseFlags() (domain string, outPath string, timeout time.Duration, asnProv string, pretty bool) {
	domainFlag := flag.String("domain", "", "Domain to recon (e.g. example.com)")
	outPathFlag := flag.String("out", "", "Output JSON path (default: ./out/<domain>.json)")
	timeoutFlag := flag.Duration("timeout", 15*time.Second, "Overall recon timeout")
	asnProvFlag := flag.String("asn-provider", "mock", "ASN provider (mock)")
	prettyFlag := flag.Bool("pretty", true, "Pretty-print JSON")
	flag.Parse()
	return *domainFlag, *outPathFlag, *timeoutFlag, *asnProvFlag, *prettyFlag
}

// setupOutputPath creates the output directory and returns the file path for results.
func setupOutputPath(op string, domain string) string {
	if op == "" {
		_ = os.MkdirAll("out", 0o755)
		op = filepath.Join("out", fmt.Sprintf("%s.json", domain))
	}
	return op
}

// getASNProvider returns the appropriate ASN provider based on the provider name.
func getASNProvider(providerName string) (modules.ASNProvider, error) {
	switch strings.ToLower(strings.TrimSpace(providerName)) {
	case "mock", "":
		return &modules.MockASNProvider{}, nil
	default:
		return nil, fmt.Errorf("unsupported -asn-provider %q (supported: mock)", providerName)
	}
}

// marshalOutput encodes the data to JSON with optional pretty-printing.
func marshalOutput(data interface{}, pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(data, "", "  ")
	}
	return json.Marshal(data)
}

func main() {
	domain, outPath, timeout, asnProv, pretty := parseFlags()

	d := normalizeDomain(domain)
	if d == "" {
		fmt.Fprintln(os.Stderr, "error: -domain is required")
		os.Exit(2)
	}

	op := setupOutputPath(outPath, d)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	agg := aggregator.New(net.DefaultResolver)

	provider, err := getASNProvider(asnProv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	out, err := agg.Collect(ctx, d, provider)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}

	b, err := marshalOutput(out, pretty)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: marshal json: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(op, append(b, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error: write output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(op)
}

func normalizeDomain(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.TrimSuffix(s, ".")
	return s
}
