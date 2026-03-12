GO ?= go
BIN_DIR ?= bin
CMD := ./cmd/magecli
SOURCES := $(shell find cmd internal pkg -name '*.go')

VERSION ?= $(shell \
	if git describe --tags --exact-match >/dev/null 2>&1; then \
		git describe --tags --exact-match; \
	else \
		short=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown"); \
		if git diff-index --quiet HEAD 2>/dev/null; then \
			echo "dev-$$short"; \
		else \
			echo "dev-$$short-dirty"; \
		fi; \
	fi \
)
COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X github.com/atlanticbt/magecli/internal/build.versionFromLdflags=$(VERSION) \
	-X github.com/atlanticbt/magecli/internal/build.commitFromLdflags=$(COMMIT) \
	-X github.com/atlanticbt/magecli/internal/build.dateFromLdflags=$(BUILD_DATE)

.PHONY: build fmt lint test tidy clean sync-skills check-skills

build: $(BIN_DIR)/magecli

sync-skills:
	@echo "Syncing skills from skills/ to .claude/skills/ and .codex/skills/..."
	@mkdir -p .claude/skills/magecli .codex/skills/magecli
	@cp -R skills/magecli/* .claude/skills/magecli/
	@cp -R skills/magecli/* .codex/skills/magecli/
	@echo "Done"

check-skills:
	@echo "Checking skill sync..."
	@diff -rq skills/magecli .claude/skills/magecli || (echo ".claude/skills/magecli out of sync" && exit 1)
	@diff -rq skills/magecli .codex/skills/magecli || (echo ".codex/skills/magecli out of sync" && exit 1)
	@echo "Skills in sync"

$(BIN_DIR)/magecli: $(SOURCES) go.mod go.sum
	@mkdir -p $(BIN_DIR)
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/magecli $(CMD)

fmt:
	$(GO) fmt ./...

lint:
	golangci-lint run

test:
	$(GO) test ./...

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BIN_DIR) dist/
