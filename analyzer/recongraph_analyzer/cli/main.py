from __future__ import annotations

import argparse
import json
import os

import networkx as nx

from recongraph_analyzer.analysis.engine import AnalysisEngine
from recongraph_analyzer.graph.builder import GraphBuilder
from recongraph_analyzer.io.record_loader import RecordLoader


class ReconGraphAnalyzerCLI:
    def __init__(
        self,
        loader: RecordLoader | None = None,
        builder: GraphBuilder | None = None,
        engine: AnalysisEngine | None = None,
    ) -> None:
        self._loader = loader or RecordLoader()
        self._builder = builder or GraphBuilder()
        self._engine = engine or AnalysisEngine(self._builder)

    def run(self, argv: list[str] | None = None) -> int:
        ap = argparse.ArgumentParser(description="ReconGraph analyzer (NetworkX)")
        ap.add_argument(
            "--input",
            required=True,
            help="Path to a recon JSON file or a directory containing JSON files",
        )
        ap.add_argument("--graph-json", default="", help="Export graph as node-link JSON")
        ap.add_argument("--graphml", default="", help="Export graph as GraphML (Gephi)")
        ap.add_argument("--top", type=int, default=10, help="Top N hubs to report")
        args = ap.parse_args(argv)

        records = self._loader.load(args.input)
        if not records:
            raise SystemExit("No records found to analyze.")

        g = self._builder.build(records)
        analysis = self._engine.analyze(g, top_n=args.top)
        self._engine.print_insights(analysis)

        if args.graph_json:
            self._export_graph_json(g, args.graph_json)
            print(f"\nWrote graph JSON to: {args.graph_json}")
        if args.graphml:
            self._export_graphml(g, args.graphml)
            print(f"Wrote GraphML to: {args.graphml}")

        return 0

    def _export_graph_json(self, g: nx.Graph, out_path: str) -> None:
        data = nx.node_link_data(g)
        os.makedirs(os.path.dirname(out_path) or ".", exist_ok=True)
        with open(out_path, "w", encoding="utf-8") as f:
            json.dump(data, f, indent=2)
            f.write("\n")

    def _export_graphml(self, g: nx.Graph, out_path: str) -> None:
        os.makedirs(os.path.dirname(out_path) or ".", exist_ok=True)
        nx.write_graphml(g, out_path)


def main(argv: list[str] | None = None) -> int:
    return ReconGraphAnalyzerCLI().run(argv)

