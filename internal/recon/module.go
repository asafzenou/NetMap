package recon

import "context"

// ModuleResult is a loosely-typed payload produced by a ReconModule.
// Concrete modules should return a strongly-typed struct and the aggregator
// will type-assert based on module Name().
type ModuleResult any

// ReconModule is the Strategy abstraction: each module implements a specific recon capability.
//
// Architectural intent:
// - The CLI orchestrator depends on this interface (dependency inversion)
// - New recon capabilities are added by implementing this interface and registering via the factory
type ReconModule interface {
	Name() string
	Run(ctx context.Context, domain string) (ModuleResult, error)
}

