"""Normalize local, checksum-verified discovery inputs without contacting a tenant."""
import argparse
import base64
import hashlib
import json
import pathlib
import re
import zipfile


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("source", type=pathlib.Path)
    parser.add_argument("--output", type=pathlib.Path, required=True,
                        help="New directory for legacy seed format; never overwrite the accepted baseline")
    args = parser.parse_args()
    source = args.source
    target = args.output
    if target.exists():
        raise ValueError("Output directory already exists; choose a new candidate path")
    target.mkdir(parents=True)
    gallery = json.loads((source / "gallery.json").read_text())
    archive = (source / "module.zip").read_bytes()
    digest = base64.b64encode(hashlib.sha512(archive).digest()).decode()
    if digest != gallery["sha512"]:
        raise ValueError("Package SHA512 does not match Gallery metadata")
    runtime = json.loads((source / "runtime.json").read_text(encoding="utf-8-sig"))
    if runtime["module_version"] != "8.1.0":
        raise ValueError("Unexpected runtime version")
    names = {name.casefold(): name for name in gallery["commands"]}
    observed = {cmd["name"].casefold(): cmd for cmd in runtime["commands"]}
    files = []
    with zipfile.ZipFile(source / "module.zip") as package:
        manifest = package.read("MicrosoftTeams.psd1").decode("utf-8-sig")
        active = "\n".join(line for line in manifest.splitlines() if not line.lstrip().startswith("#"))
        candidates = set(re.findall(r"'([A-Z][A-Za-z]+-[A-Za-z][A-Za-z0-9]+)'", active))
        for name in candidates:
            names.setdefault(name.casefold(), name)
        for path in sorted(package.namelist()):
            if path.endswith((".psd1", ".psm1", ".ps1", "Help.xml", "help.xml", "CmdletSettings.json", "CmdletConfig.json")):
                files.append({"path": path, "sha256": hashlib.sha256(package.read(path)).hexdigest()})
    for key, cmd in observed.items():
        names.setdefault(key, cmd["name"])
    gallery_keys = {name.casefold() for name in gallery["commands"]}
    manifest_keys = {name.casefold() for name in candidates}
    commands = []
    for key, name in sorted(names.items()):
        cmd = observed.get(key)
        commands.append({"name": name, "noun": name.split("-", 1)[1], "gallery": key in gallery_keys,
                         "manifest_candidate": key in manifest_keys, "runtime_observed": cmd is not None,
                         "runtime_name": cmd["name"] if cmd else None,
                         "parameter_sets": len(cmd["parameter_sets"]) if cmd else None,
                         "qualification": "metadata-only", "implementation": "not-started"})
    metadata = {"schema_version": 1, "module_version": "8.1.0", "package_url": "https://www.powershellgallery.com/api/v2/package/MicrosoftTeams/8.1.0",
                "package_sha512_base64": digest, "package_sha256": hashlib.sha256(archive).hexdigest(),
                "gallery_count": len(gallery_keys), "runtime_count": len(observed), "candidate_count": len(commands),
                "runtime_environment": {k: v for k, v in runtime.items() if k != "commands"},
                "limitations": ["Unauthenticated macOS observation only; conditional/platform exports require reconciliation.",
                                "Manifest candidates use lexical extraction, not authoritative PowerShell export evaluation.",
                                "Runtime metadata does not establish wire contracts, default values, server behaviour or complete output schemas.",
                                "Validation extraction covers named public properties; custom validators require source/behaviour research."],
                "source_files": files, "release_notes": gallery["release_notes"]}
    (target / "provenance.json").write_text(json.dumps(metadata, indent=2) + "\n")
    (target / "commands.json").write_text(json.dumps(commands, indent=2) + "\n")
    # One command per line keeps the full parameter-set evidence reviewable without
    # repeating indentation across hundreds of common PowerShell parameters.
    (target / "runtime.json").write_text('{"environment":' + json.dumps(metadata["runtime_environment"]) + ',"commands":[\n' + ',\n'.join(json.dumps(c, separators=(",", ":")) for c in runtime["commands"]) + '\n]}\n')
    print(f"Normalized {len(commands)} candidates; {len(observed)} runtime exports")


if __name__ == "__main__":
    main()
