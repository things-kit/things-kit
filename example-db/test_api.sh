#!/bin/bash
# Manual test script for example-db API

set -e

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Example-DB API Manual Test ===${NC}\n"

# Health check
echo -e "${GREEN}1. Health Check${NC}"
curl -s $BASE_URL/health | jq .
echo -e "\n"

# Create user 1
echo -e "${GREEN}2. Create User - John Doe${NC}"
curl -s -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com"}' | jq .
echo -e "\n"

# Create user 2
echo -e "${GREEN}3. Create User - Jane Smith${NC}"
curl -s -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane Smith","email":"jane@example.com"}' | jq .
echo -e "\n"

# Create user 3
echo -e "${GREEN}4. Create User - Bob Johnson${NC}"
curl -s -X POST $BASE_URL/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Bob Johnson","email":"bob@example.com"}' | jq .
echo -e "\n"

# List all users
echo -e "${GREEN}5. List All Users${NC}"
curl -s $BASE_URL/users | jq .
echo -e "\n"

# Get user by ID
echo -e "${GREEN}6. Get User by ID (ID=1)${NC}"
curl -s $BASE_URL/users/1 | jq .
echo -e "\n"

# Update user
echo -e "${GREEN}7. Update User (ID=1)${NC}"
curl -s -X PUT $BASE_URL/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"John Updated","email":"john.updated@example.com"}' | jq .
echo -e "\n"

# Get updated user
echo -e "${GREEN}8. Get Updated User (ID=1)${NC}"
curl -s $BASE_URL/users/1 | jq .
echo -e "\n"

# Delete user
echo -e "${GREEN}9. Delete User (ID=2)${NC}"
curl -s -X DELETE $BASE_URL/users/2 -w "\nStatus: %{http_code}\n"
echo -e "\n"

# List users after deletion
echo -e "${GREEN}10. List Users After Deletion${NC}"
curl -s $BASE_URL/users | jq .
echo -e "\n"

echo -e "${BLUE}=== All tests completed! ===${NC}"
