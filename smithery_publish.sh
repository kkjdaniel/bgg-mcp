#!/bin/bash

# Publishes the MCP bundle to Smithery (https://smithery.ai/servers/kkjdaniel/bgg-mcp)
# as a stdio release.
#
# `smithery mcp publish` can't be used: it copies the manifest's tools into the
# server card, and Smithery rejects tools without an inputSchema, which the MCPB
# manifest schema doesn't allow. So this asks a built server for its tools,
# prompts and resources and sends them to the publish API with the bundle.
#
# Usage: SMITHERY_API_KEY=... ./smithery_publish.sh <server binary> [bundle]
# The binary can be built for any platform; it only answers the list requests.

set -euo pipefail

SERVER_BIN=${1:?Usage: SMITHERY_API_KEY=... $0 <server binary> [bundle]}
BUNDLE=${2:-build/bgg-mcp.mcpb}
QUALIFIED_NAME="kkjdaniel/bgg-mcp"
API="https://api.smithery.ai/servers/${QUALIFIED_NAME/\//%2F}/releases"

for cmd in jq curl unzip; do
    if ! command -v "$cmd" &> /dev/null; then
        echo "✗ Error: $cmd is not installed"
        exit 1
    fi
done

if [ -z "${SMITHERY_API_KEY:-}" ]; then
    echo "✗ Error: SMITHERY_API_KEY is not set"
    exit 1
fi

if [ ! -f "$BUNDLE" ]; then
    echo "✗ Error: $BUNDLE not found. Run make mcpb first"
    exit 1
fi

# One stdio session: the server answers each request before it sees stdin close
RESPONSES=$(printf '%s\n' \
    '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"smithery-publish","version":"1.0.0"}}}' \
    '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
    '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' \
    '{"jsonrpc":"2.0","id":3,"method":"prompts/list"}' \
    '{"jsonrpc":"2.0","id":4,"method":"resources/list"}' \
    | BGG_DISABLE_UPDATE_CHECK=true "$SERVER_BIN" 2> /dev/null)

MANIFEST=$(unzip -p "$BUNDLE" manifest.json)

# configSchema mirrors the CLI's conversion of the manifest's user_config
PAYLOAD=$(jq -s --argjson manifest "$MANIFEST" '
    (map({key: (.id | tostring), value: .result}) | from_entries) as $r
    | if ($r | has("1") and has("2") and has("3") and has("4")) | not
      then error("the server did not answer every list request") else . end
    | if $r["1"].serverInfo.version != $manifest.version
      then error("server version \($r["1"].serverInfo.version) does not match bundle version \($manifest.version)") else . end
    | {
        type: "stdio",
        runtime: "binary",
        serverCard: {
          serverInfo: {
            name: $r["1"].serverInfo.name,
            title: $manifest.display_name,
            version: $r["1"].serverInfo.version,
            description: $manifest.description,
            websiteUrl: $manifest.homepage
          },
          tools: $r["2"].tools,
          prompts: $r["3"].prompts,
          resources: $r["4"].resources
        },
        configSchema: {
          type: "object",
          properties: ($manifest.user_config | map_values({type, title, description, default}))
        }
      }
' <<< "$RESPONSES")

echo "Publishing $QUALIFIED_NAME $(jq -r '.serverCard.serverInfo.version' <<< "$PAYLOAD") to Smithery" \
    "(tools: $(jq '.serverCard.tools | length' <<< "$PAYLOAD")," \
    "prompts: $(jq '.serverCard.prompts | length' <<< "$PAYLOAD")," \
    "resources: $(jq '.serverCard.resources | length' <<< "$PAYLOAD"))"

# --form-string, as -F would read a ";" in the JSON as a field option
RELEASE=$(curl -sS --fail-with-body -X PUT "$API" \
    -H "Authorization: Bearer $SMITHERY_API_KEY" \
    --form-string "payload=$PAYLOAD" \
    -F "bundle=@$BUNDLE;type=application/octet-stream")

RELEASE_ID=$(jq -r '.deploymentId' <<< "$RELEASE")
STATUS=$(jq -r '.status' <<< "$RELEASE")
jq -r '.warnings[]? | "! \(.)"' <<< "$RELEASE"
echo "Release $RELEASE_ID accepted"

# Wait up to 10 minutes for the release to finish, printing its logs as they arrive
SHOWN=0
for _ in $(seq 1 120); do
    case "$STATUS" in
        SUCCESS)
            echo "✓ Published to https://smithery.ai/servers/$QUALIFIED_NAME"
            exit 0
            ;;
        FAILURE|FAILURE_SCAN|INTERNAL_ERROR|CANCELLED|AUTH_REQUIRED)
            echo "✗ Release $RELEASE_ID ended with status $STATUS"
            exit 1
            ;;
    esac
    sleep 5
    INFO=$(curl -sS --fail-with-body "$API/$RELEASE_ID" -H "Authorization: Bearer $SMITHERY_API_KEY")
    STATUS=$(jq -r '.status' <<< "$INFO")
    jq -r --argjson shown "$SHOWN" '.logs // [] | .[$shown:][] | "[\(.stage)] \(.message)\(if .error then ": \(.error.message)" else "" end)"' <<< "$INFO"
    SHOWN=$(jq '.logs // [] | length' <<< "$INFO")
done

echo "✗ Release $RELEASE_ID still $STATUS after 10 minutes. See https://smithery.ai/servers/$QUALIFIED_NAME/releases"
exit 1
