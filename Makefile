.PHONY: build clean all mcpb

MCPB_CLI := npx -y @anthropic-ai/mcpb@2.1.2
MCPB_STAGE := build/mcpb
RELEASE_FLAGS := CGO_ENABLED=0
RELEASE_BUILD := go build -trimpath -ldflags="-s -w"

build:
	mkdir -p build
	go build -o build/bgg-mcp

clean:
	rm -rf build/

all: clean build

# Packs build/bgg-mcp.mcpb with a universal macOS binary and a Windows binary.
# Needs macOS (lipo, sips) and Node.js.
mcpb:
	rm -rf $(MCPB_STAGE)
	mkdir -p $(MCPB_STAGE)/server
	$(RELEASE_FLAGS) GOOS=darwin GOARCH=arm64 $(RELEASE_BUILD) -o build/bgg-mcp-darwin-arm64
	$(RELEASE_FLAGS) GOOS=darwin GOARCH=amd64 $(RELEASE_BUILD) -o build/bgg-mcp-darwin-amd64
	lipo -create -output $(MCPB_STAGE)/server/bgg-mcp build/bgg-mcp-darwin-arm64 build/bgg-mcp-darwin-amd64
	$(RELEASE_FLAGS) GOOS=windows GOARCH=amd64 $(RELEASE_BUILD) -o $(MCPB_STAGE)/server/bgg-mcp.exe
	cp mcpb/manifest.json $(MCPB_STAGE)/manifest.json
	sips -z 512 512 images/bgg-mcp-logo.png --out $(MCPB_STAGE)/icon.png
	$(MCPB_CLI) pack $(MCPB_STAGE) build/bgg-mcp.mcpb
