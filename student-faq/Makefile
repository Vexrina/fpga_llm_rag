.PHONY: help install dev build lint typecheck clean run

help:
	@echo "Available commands:"
	@echo "  make install      - Install dependencies"
	@echo "  make dev          - Run development server"
	@echo "  make build        - Build for production"
	@echo "  make lint         - Run linter"
	@echo "  make typecheck    - Run TypeScript type check"
	@echo "  make clean        - Remove node_modules and build artifacts"
	@echo "  make run          - Alias for make dev"

install:
	npm install

dev:
	npm run dev

build:
	npm run build

lint:
	npm run lint

typecheck:
	npm run typecheck

clean:
	rm -rf node_modules dist

run: dev
