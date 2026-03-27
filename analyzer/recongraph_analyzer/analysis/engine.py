from __future__ import annotations

from dataclasses import dataclass
from typing import Any, Dict, List, Tuple

import networkx as nx

from recongraph_analyzer.graph.builder import GraphBuilder


@dataclass(frozen=True)
class GraphAnalysis:
    meta: Dict[str, Any]
    degree_centrality_top: List[Tuple[str, float]]
    shared_dependencies: Dict[str, List[Tuple[str, int]]]


class AnalysisEngine:
    def __init__(self, builder: GraphBuilder | None = None) -> None:
        self._builder = builder or GraphBuilder()

    def analyze(self, g: nx.Graph, top_n: int = 10) -> GraphAnalysis:
        degree_centrality = nx.degree_centrality(g) if g.number_of_nodes() else {}
        hubs: List[Tuple[str, float]] = sorted(
            degree_centrality.items(), key=lambda x: x[1], reverse=True
        )[:top_n]

        components = list(nx.connected_components(g))
        components_sorted = sorted(components, key=len, reverse=True)

        shared = self._builder.summarize_shared_dependencies(g)

        meta = {
            "nodes": g.number_of_nodes(),
            "edges": g.number_of_edges(),
            "connected_components": len(components_sorted),
            "largest_component_size": len(components_sorted[0]) if components_sorted else 0,
        }

        return GraphAnalysis(meta=meta, degree_centrality_top=hubs, shared_dependencies=shared)

    def print_insights(self, analysis: GraphAnalysis, max_lines: int = 20) -> None:
        m = analysis.meta
        print(
            f"Graph: {m.get('nodes', 0)} nodes, {m.get('edges', 0)} edges, "
            f"{m.get('connected_components', 0)} components "
            f"(largest={m.get('largest_component_size', 0)})"
        )

        print("\nTop hubs (degree centrality):")
        for node_id, score in analysis.degree_centrality_top[:10]:
            print(f"  - {node_id}  score={score:.4f}")

        for kind, label in (("ns", "nameservers"), ("ip", "IPs"), ("asn", "ASNs")):
            items = analysis.shared_dependencies.get(kind, [])
            if not items:
                continue
            print(f"\nShared {label}:")
            for value, n in items[:max_lines]:
                if kind == "asn":
                    print(f"  - {value} is a dependency for {n} IPs")
                else:
                    print(f"  - {n} domains share {kind}={value}")

