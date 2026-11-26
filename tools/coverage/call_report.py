#!/usr/bin/env python3
"""Render the CALL command/parameter-set inventory without promoting evidence."""

import argparse
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
DESTINATION = ROOT / "docs/implementation/calls/coverage-report.md"
FAMILIES = {
    "attendant_builders": "auto-attendants-contract-notes.md",
    "audio_prompts": "audio-prompts-contract-notes.md",
    "auto_attendants": "auto-attendants-contract-notes.md",
    "call_queues": "call-queues-contract-notes.md",
    "call_templates": "call-templates-contract-notes.md",
    "compliance_applications": "compliance-recording-contract-notes.md",
    "mainline_attendants": "mainline-attendants-contract-notes.md",
    "resource_accounts": "resource-accounts-contract-notes.md",
    "schedules": "schedules-contract-notes.md",
}


def reference(value):
    """Link the declared symbol, not an inferred method or test."""
    return f"[`{value['name']}`](../../../{value['file']})"


def render(ledger):
    records = sorted(
        (row for row in ledger["records"] if row["resource"] in FAMILIES),
        key=lambda row: (row["resource"], row["name"].casefold()),
    )
    if {row["resource"] for row in records} != FAMILIES.keys():
        raise ValueError("CALL family missing from the coverage ledger")
    if any(not (row["implementation_owner"] or "").startswith("CALL-") for row in records):
        raise ValueError("CALL resource has a foreign or missing implementation owner")
    sets = sum(len(row["parameter_sets"]) for row in records)
    lines = [
        "# CALL command and parameter-set evidence inventory",
        "",
        "Owner: CALL-90 (#188). Baseline: MicrosoftTeams 8.1.0.",
        "",
        f"This inventory names {len(records)} commands and {sets} observed parameter sets",
        "across all nine CALL families. Regenerate with",
        "`python3 tools/coverage/call_report.py`; use `--check` to detect stale output.",
        "",
        "## Interpretation",
        "",
        "Symbol and test links are declared mappings, not proof that those tests ran",
        "or that they cover every behavior. States and limitations below come from",
        "the current ledger; this report never promotes them. A synthetic success",
        "body is offline model evidence, not a successful tenant operation.",
        "",
        "The integrated offline suite is",
        "[`tests/integration/lab/call_integration`](../../../tests/integration/lab/call_integration/contracts_test.go).",
        "It exercises schedule/association identities, audio IDs across prompt/queue",
        "models, and independent recovery after a different service denies access.",
        "The four offline composed scenarios are synthetic. The separate opt-in",
        "`TestAccScheduleAutoAttendantComposition` journals owned weekly/fixed schedules and an",
        "attendant, verifies populated holiday read plus a nonempty export receipt, and the associations before/after",
        "replacement against pinned PowerShell, and independently removes all three.",
        "Registration is not a live pass; serialized export-content semantics and holiday import remain unqualified.",
        "Existing child fixtures retain their provenance. Executed results and the tested",
        "revision belong in the implementation PR/CI, separate from this inventory.",
        "",
        "## Recorded scoped tenant result",
        "",
        "[Run 45874462217](https://github.com/deploymenttheory/go-sdk-msteams/actions/runs/45874462217)",
        "passed the original owned weekly-schedule/attendant composition using client-secret",
        "authentication. The [sanitized evidence](evidence/schedule-attendant-f1931de.json)",
        "records CI merge revision `866e6e23d28c7d8794a26342f3c5e5825c7e0af3`,",
        "source head `f1931de0309afbf45353647922212d56b2419c0d`, four asserted phases,",
        "two pinned comparisons and two cleaned journal entries with zero unresolved.",
        "This pass predates the fixed-holiday/read/export extension and does not",
        "establish a live pass for that extension or full command/parameter-set coverage.",
        "",
        "The [holiday extension run](evidence/schedule-attendant-0b92810-failed.json)",
        "at source head `0b92810` failed during attendant creation; both schedules were",
        "cleaned but the planned attendant receipt remains unresolved. Independent",
        "journal-bound SDK and pinned-module exact-name paginated lookups both found",
        "current absence; neither lookup retroactively resolves the lost receipt.",
        "The [server-normalized retry](evidence/schedule-attendant-0ab716c-failed.json)",
        "at source head `0ab716c` also failed at creation, now with HTTP 400 and a",
        "bounded duplicate-reference category. Its three entries independently cleaned;",
        "this proves recovery for that rejection, not a successful holiday composition.",
        "The [distinct-flow run](evidence/schedule-attendant-542da26-failed.json) then",
        "created the attendant successfully but failed the holiday visualization assertion;",
        "all three entries independently cleaned. The [pinned projection](evidence/holiday-date-projection-8.1.0.json)",
        "requires day-first invariant reply dates, distinct from ISO request strings.",
        "The next run retains exact calendar boundaries and captures only sanitized holiday JSON.",
        "",
        "",
        "## Qualification boundaries",
        "",
        "- Resource-account successful creation/association needs directory-write or",
        "  delegated credentials unavailable to the existing app-only lab identity.",
        "- Compliance recording application success needs a real resource account;",
        "  grants affect calling behavior and are not qualified by policy CRUD.",
        "- Auto-recording template success needs SharePoint site-admin access; an",
        "  expected denial is not a successful create/get/update/delete lifecycle.",
        "- Schedule/attendant references, audio prompt use and populated resource",
        "  links require live compositions beyond the isolated child lifecycles.",
        "- Commercial application-auth evidence does not qualify delegated flows,",
        "  sovereign clouds or effective calling behavior.",
        "",
        "CALL-90 remains incomplete until the integrated acceptance and complete",
        "contract criteria are demonstrated. Child issue closure and mapped symbols",
        "alone do not sign off CALL-90 or the CALL epic.",
        "",
    ]
    for family, notes in FAMILIES.items():
        lines += [f"## {family}", "", f"Contract and live limitations: [{notes}]({notes}).", ""]
        for row in (row for row in records if row["resource"] == family):
            lines += [f"### {row['name']}", "", f"Owner: `{row['implementation_owner']}`. "
                      f"Implementation: `{row['implementation']['status']}`; "
                      f"contract: `{row['contract']['status']}`; live: `{row['live']['status']}`.", "",
                      "| Parameter set | Go symbols | Contract test mappings | Live test mappings |",
                      "| --- | --- | --- | --- |"]
            for parameter_set in row["parameter_sets"]:
                cells = [f"`{parameter_set['name']}`"]
                for field in ("symbols", "contract_tests", "live_tests"):
                    cells.append("<br>".join(reference(value) for value in parameter_set[field]) or "None recorded")
                lines.append("| " + " | ".join(cells) + " |")
            if not row["parameter_sets"]:
                lines.append("| No observed parameter set | Unresolved | Unresolved | Unresolved |")
            lines += ["", "Live limitation: " + row["live"].get("reason", "No limitation recorded."), ""]
    return "\n".join(lines)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="fail if the committed report is stale")
    args = parser.parse_args()
    rendered = render(json.loads((ROOT / "contracts/coverage/8.1.0.json").read_text()))
    if args.check:
        if not DESTINATION.exists() or DESTINATION.read_text() != rendered:
            parser.exit(1, "CALL coverage report is stale; run tools/coverage/call_report.py\n")
        print("CALL coverage report matches the current ledger")
    else:
        DESTINATION.write_text(rendered)


if __name__ == "__main__":
    main()
