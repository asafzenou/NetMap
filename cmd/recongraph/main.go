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

	"recongraph/internal/aggregator"
	"recongraph/internal/modules/asn"
)

func main() {
	var (
		domain   = flag.String("domain", "", "Domain to recon (e.g. example.com)")
		outPath  = flag.String("out", "", "Output JSON path (default: ./out/<domain>.json)")
		timeout  = flag.Duration("timeout", 15*time.Second, "Overall recon timeout")
		asnProv  = flag.String("asn-provider", "mock", "ASN provider (mock)")
		pretty   = flag.Bool("pretty", true, "Pretty-print JSON")
	)
	flag.Parse()

	d := normalizeDomain(*domain)
	if d == "" {
		fmt.Fprintln(os.Stderr, "error: -domain is required")
		os.Exit(2)
	}

	op := *outPath
	if op == "" {
		_ = os.MkdirAll("out", 0o755)
		op = filepath.Join("out", fmt.Sprintf("%s.json", d))
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	agg := aggregator.New(net.DefaultResolver)

	var provider asn.Provider
	switch strings.ToLower(strings.TrimSpace(*asnProv)) {
	case "mock", "":
		provider = &asn.MockProvider{}
	default:
		fmt.Fprintf(os.Stderr, "error: unsupported -asn-provider %q (supported: mock)\n", *asnProv)
		os.Exit(2)
	}

	out, err := agg.Collect(ctx, d, provider)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}

	var b []byte
	if *pretty {
		b, err = json.MarshalIndent(out, "", "  ")
	} else {
		b, err = json.Marshal(out)
	}
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

