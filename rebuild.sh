#!/bin/bash
echo "🚀 Rebuilding ModernFM from Scratch..."
sudo docker compose -f deploy/docker-compose.yml down
sudo docker system prune -f
sudo docker compose -f deploy/docker-compose.yml up -d --build
echo "✅ Done. Access UI at http://localhost:38866"
