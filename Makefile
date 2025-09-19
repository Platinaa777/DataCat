GO ?= go
GO_TEST_FLAGS ?=

.PHONY: build run test test-race test-integration vet fmt ui ui-check dev generate

build:
	$(GO) build -trimpath -o bin/datacat ./cmd/datacat
	$(GO) build -trimpath -o bin/import-legacy ./cmd/import-legacy

run:
	$(GO) run ./cmd/datacat -migrate

test:
	$(GO) test $(GO_TEST_FLAGS) ./...

test-race:
	$(GO) test $(GO_TEST_FLAGS) -race ./...

test-integration:
	@test -n "$(TEST_DATABASE_URL)" || (echo 'Set TEST_DATABASE_URL to a dedicated database ending in _test' >&2; exit 1)
	$(GO) test $(GO_TEST_FLAGS) -tags integration ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -w cmd internal pkg api/gen

ui:
	cd frontend/datacat-ui && npm start

ui-check:
	cd frontend/datacat-ui && npm run lint && npm run typecheck && npm run build

dev:
	./scripts/dev.sh

generate:
	./scripts/generate-proto.sh
