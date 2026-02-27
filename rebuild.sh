#!/bin/bash
echo "🚀 Rebuilding ModernFM from Scratch..."
docker-compose -f deploy/docker-compose.yml down
docker system prune -f
docker-compose -f deploy/docker-compose.yml up -d --build
echo "✅ Done. Access UI at http://localhost:38866"
