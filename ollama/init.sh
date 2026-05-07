#!/bin/bash
set -e

echo "Pulling embedding models..."
ollama pull bge-m3 || true
ollama pull nomic-embed-text-v2-moe || true

echo "Starting Ollama server..."
exec ollama serve