from __future__ import annotations

from dataclasses import dataclass
from typing import Dict, Iterable, List, Tuple

import networkx as nx

from recongraph_analyzer.io.record_loader import ReconRecord


@dataclass(frozen=True)
class NodeKey:
    kind: str
    value: str

    def id(self) -> str:
        return f"{self.kind}:{self.value}"


class GraphBuilder:
    """
    Builds an undirected infrastructure dependency graph:
      - domain -- ip
      - domain -- ns
      - ip -- asn
    """

    def build(self, records: Iterable[ReconRecord]) -> nx.Graph:
        g = nx.Graph()

        for r in records:
            if not r.domain:
                continue

            d = NodeKey("domain", r.domain)
            self._add_node(g, d, label=r.domain, kind="domain")

            for ip in r.ips:
                n = NodeKey("ip", ip)
                self._add_node(g, n, label=ip, kind="ip")
                g.add_edge(d.id(), n.id(), kind="domain_ip")

                if r.asn:
                    a = NodeKey("asn", r.asn)
                    self._add_node(g, a, label=r.asn, kind="asn")
                    g.add_edge(n.id(), a.id(), kind="ip_asn")

            for ns in r.ns:
                n = NodeKey("ns", ns)
                self._add_node(g, n, label=ns, kind="ns")
                g.add_edge(d.id(), n.id(), kind="domain_ns")

        return g

    def summarize_shared_dependencies(self, g: nx.Graph) -> Dict[str, List[Tuple[str, int]]]:
        """
        Rank shared dependencies:
          - ns: number of domains connected to each nameserver
          - ip: number of domains connected to each IP
          - asn: number of IPs connected to each ASN
        """
        out: Dict[str, List[Tuple[str, int]]] = {"ns": [], "ip": [], "asn": []}

        for kind in ("ns", "ip"):
            counts: List[Tuple[str, int]] = []
            for node_id, data in g.nodes(data=True):
                if data.get("kind") != kind:
                    continue
                domains = [
                    nbr
                    for nbr in g.neighbors(node_id)
                    if g.nodes[nbr].get("kind") == "domain"
                ]
                if len(domains) > 1:
                    value = node_id.split(":", 1)[1]
                    counts.append((value, len(domains)))
            out[kind] = sorted(counts, key=lambda x: (-x[1], x[0]))

        counts_asn: List[Tuple[str, int]] = []
        for node_id, data in g.nodes(data=True):
            if data.get("kind") != "asn":
                continue
            ips = [nbr for nbr in g.neighbors(node_id) if g.nodes[nbr].get("kind") == "ip"]
            if len(ips) > 1:
                value = node_id.split(":", 1)[1]
                counts_asn.append((value, len(ips)))
        out["asn"] = sorted(counts_asn, key=lambda x: (-x[1], x[0]))

        return out

    def _add_node(self, g: nx.Graph, node: NodeKey, **attrs) -> None:
        node_id = node.id()
        if node_id not in g:
            g.add_node(node_id, **attrs)

