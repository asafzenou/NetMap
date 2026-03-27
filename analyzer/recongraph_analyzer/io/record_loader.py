from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any, Dict, Iterable, Iterator, List, Optional


@dataclass(frozen=True)
class ReconRecord:
    domain: str
    ips: List[str]
    ns: List[str]
    asn: str
    raw: Dict[str, Any]

    @staticmethod
    def from_dict(d: Dict[str, Any]) -> "ReconRecord":
        domain = str((d.get("domain") or "")).strip().lower().rstrip(".")
        ips = [str(x).strip() for x in _as_list(d.get("ips")) if str(x).strip()]
        ns = [
            str(x).strip().lower().rstrip(".")
            for x in _as_list(d.get("ns"))
            if str(x).strip()
        ]
        asn = str((d.get("asn") or "")).strip().upper()
        return ReconRecord(domain=domain, ips=ips, ns=ns, asn=asn, raw=d)


class RecordLoader:
    def load(self, input_path: str) -> List[ReconRecord]:
        return list(self.iter_load(input_path))

    def iter_load(self, input_path: str) -> Iterator[ReconRecord]:
        p = Path(input_path)
        if p.is_dir():
            for fp in sorted(p.glob("*.json")):
                yield from self.iter_load(str(fp))
            return

        if not p.is_file():
            return

        with p.open("r", encoding="utf-8") as f:
            data = json.load(f)
        rec = ReconRecord.from_dict(data)
        if rec.domain:
            yield rec


def _as_list(v: Any) -> List[Any]:
    if v is None:
        return []
    if isinstance(v, list):
        return v
    return [v]

