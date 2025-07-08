#!/usr/bin/env python3

# A simple proxy for GitHub Copilot's MCP API.
# allows tools such as gemini-cli to use a remote MCP servers.
# It act similar to a local MCP server,

import sys
import os
import json
import requests

token = os.environ.get("GITHUB_PERSONAL_ACCESS_TOKEN")

if not token:
    # If the token is missing, send an error back in the JSON-RPC format
    error_response = {
        "jsonrpc": "2.0",
        "id": 1,
        "error": {"code": -32000, "message": "GITHUB_PERSONAL_ACCESS_TOKEN not set"},
    }
    print(json.dumps(error_response))
    sys.stdout.flush()
    sys.exit(1)

headers = {
    "Authorization": f"Bearer {token}",
    "Content-Type": "application/json",
    "Accept": "application/json",
}

remote_url = "https://api.githubcopilot.com/mcp/"

# Use a session to keep the connection alive
session = requests.Session()

for line in sys.stdin:
    try:
        request_data = json.loads(line)

        response = session.post(remote_url, json=request_data, headers=headers)
        response.raise_for_status()

        print(response.text, end="", flush=True)

    except json.JSONDecodeError:
        pass
    except requests.exceptions.RequestException as e:
        error_msg = {
            "jsonrpc": "2.0",
            "id": request_data.get("id", None),
            "error": {"code": -32001, "message": f"Network error: {e}"},
        }
        print(json.dumps(error_msg), flush=True)
