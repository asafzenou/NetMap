package models

import "time"

// ReconOutput is the stable JSON contract between the Go recon engine and the Python analyzer.
// Keep this backward-compatible as the system evolves.
type ReconOutput struct {
	Domain string   `json:"domain"`
	IPs    []string `json:"ips,omitempty"`
	NS     []string `json:"ns,omitempty"`
	ASN    string   `json:"asn,omitempty"`

	Metadata Metadata `json:"metadata,omitempty"`
}

type Metadata struct {
	CollectedAt time.Time `json:"collected_at"`
	DurationMs  int64     `json:"duration_ms"`
	Sources     []string  `json:"sources,omitempty"`
}

