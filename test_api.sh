#!/bin/bash
echo "Testing Backend API List..."
curl -s "http://localhost:38866/api/files/list" | head -c 200
echo -e "\n--- API Test Completed ---"
