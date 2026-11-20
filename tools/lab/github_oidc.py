#!/usr/bin/env python3
"""Write an ephemeral GitHub OIDC assertion for the opt-in lab workflow."""
import json
import os
from pathlib import Path
import sys
import tempfile
import urllib.parse
import urllib.request


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, *args, **kwargs):
        return None


def main():
    if len(sys.argv) == 2 and sys.argv[1] == "--cleanup":
        value = os.environ.get("MSTEAMS_ASSERTION_FILE")
        if value:
            target = Path(value).resolve()
            if target.parent != Path(os.environ["RUNNER_TEMP"]).resolve():
                raise ValueError("unexpected assertion path")
            target.unlink(missing_ok=True)
        return
    url = urllib.parse.urlsplit(os.environ["ACTIONS_ID_TOKEN_REQUEST_URL"])
    if (url.scheme != "https" or not url.hostname
            or not url.hostname.endswith(".actions.githubusercontent.com")
            or url.username or url.password or url.fragment or url.port not in (None, 443)):
        raise ValueError("unexpected assertion authority")
    query = urllib.parse.parse_qs(url.query, strict_parsing=True)
    query["audience"] = ["api://AzureADTokenExchange"]
    destination = urllib.parse.urlunsplit(url._replace(query=urllib.parse.urlencode(query, doseq=True)))
    request = urllib.request.Request(destination, headers={
        "Authorization": "Bearer " + os.environ["ACTIONS_ID_TOKEN_REQUEST_TOKEN"],
        "Accept": "application/json",
    })
    with urllib.request.build_opener(NoRedirect()).open(request, timeout=30) as response:
        raw = response.read((1 << 20) + 1)
    if len(raw) > 1 << 20:
        raise ValueError("assertion response too large")
    token = json.loads(raw).get("value")
    if not isinstance(token, str) or not token or "\n" in token or "\r" in token:
        raise ValueError("invalid assertion response")
    descriptor, path = tempfile.mkstemp(prefix="msteams-assertion-", suffix=".jwt", dir=os.environ["RUNNER_TEMP"])
    with os.fdopen(descriptor, "w", encoding="utf-8") as output:
        output.write(token)
    with open(os.environ["GITHUB_ENV"], "a", encoding="utf-8") as environment:
        environment.write("MSTEAMS_ASSERTION_FILE=" + path + "\n")
    print("Ephemeral OIDC assertion prepared; material is not logged")


if __name__ == "__main__":
    try:
        main()
    except Exception:
        # urllib diagnostics can contain request URLs; never emit exception text.
        print("OIDC assertion preparation or cleanup failed", file=sys.stderr)
        sys.exit(1)
