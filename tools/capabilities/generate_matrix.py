#!/usr/bin/env python3
"""Regenerate the dated operation-mode documentation overlay on the pinned command union."""

import argparse
import fnmatch
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
TARGET = ROOT / "msteams/capabilities/capabilities_8_1_0.json"
ALL_CLOUDS = ["commercial", "gcc", "gcch", "dod"]


def build():
    baseline = json.loads((ROOT / "contracts/powershell/8.1.0/reconciliation.json").read_text())
    patterns = [
        "New-Team", "Set-CsOnlineApplicationInstance", "*PolicyPackage*",
        "*-CsTeamsShiftsConnection*", "*-CsBatchTeamsDeployment*", "Get-MultiGeoRegion",
        "Get-AllM365TeamsApps", "Get-M365TeamsApp", "Update-M365TeamsApp",
        "Get-M365UnifiedCustomPendingApps", "Get-M365UnifiedTenantSettings",
        "Update-M365UnifiedTenantSettings", "Update-M365UnifiedCustomPendingApp",
    ]
    rules = [{"id": f"app-exception-{i + 1:02}", "pattern": pattern, "clouds": ALL_CLOUDS}
             for i, pattern in enumerate(patterns)]
    for verb in ["Get", "New", "Sync"]:
        rules.append({"id": f"sovereign-application-instance-{verb.lower()}",
                      "pattern": f"{verb}-CsOnlineApplicationInstance", "clouds": ["gcch", "dod"]})
    records = []
    for record in baseline["records"]:
        records.append({
            "name": record["name"], "disposition": record["disposition"],
            "parameter_sets": record["parameter_set_names"],
            "current_restrictions": [rule["id"] for rule in rules if fnmatch.fnmatchcase(
                record["name"].lower(), rule["pattern"].lower())],
        })
    return {
        "baseline_version": baseline["module_version"],
        "package_sha256": baseline["package_sha256"],
        "documentation_url": "https://learn.microsoft.com/en-us/microsoftteams/teams-powershell-application-authentication",
        "documentation_reviewed": "2026-11-20",
        "rules": rules, "records": records,
    }


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    content = json.dumps(build(), indent=2, ensure_ascii=False) + "\n"
    if args.check:
        if TARGET.read_text(encoding="utf-8") != content:
            parser.error("capability matrix differs from pinned inventory/documentation overlay")
    else:
        TARGET.write_text(content, encoding="utf-8")
    print("Capability matrix: 8.1.0 inventory and dated documentation overlay verified")
