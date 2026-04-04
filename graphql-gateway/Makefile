.PHONY: build run clean test gg gen

GQLGEN := $(shell which gqlgen 2>/dev/null || echo "$(HOME)/go/bin/gqlgen")

export PATH := $(HOME)/go1.25/bin:$(HOME)/go/bin:$(PATH)
export GOROOT := $(HOME)/go1.25

build:
	go build -o bin/graphql-gateway ./cmd/server

run: build
	./bin/graphql-gateway

clean:
	rm -rf bin/

test:
	go test -v ./...

gg gen:
	@echo "=== GraphQL Schema Generator (gqlgen) ==="
	@echo ""
	@echo "Loading GraphQL schema files:"
	@for f in graph/types/*.graphqls graph/query/*.graphqls graph/mutation/*.graphqls graph/enums/*.graphqls; do \
		if [ -f "$$f" ]; then \
			echo "  - $$f"; \
		fi \
	done
	@echo ""
	@echo "Running gqlgen generate..."
	@$(GQLGEN) generate
	@echo ""
	@echo "Generated files:"
	@echo "  - internal/app/generated/generated.go (core)"
	@echo "  - internal/app/generated/models_gen.go (models)"
	@echo "  - graph/generated_stubs/*.resolver.go (stubs with panics)"
	@echo ""
	@echo "Your resolver implementation: internal/app/resolver.go"
	@echo "New stubs with panics: graph/generated_stubs/*.resolver.go"
	@echo "Copy new stubs from graph/generated_stubs/ to internal/app/ if needed"
