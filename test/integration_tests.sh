#!/bin/bash

# Base URL for the server
BASE_URL="http://localhost:8080"

# Test Ethereum addresses
ADDRESS_1="0xdc9dcb1f7979ff6cbf3d11c46e210125e7f86bf0"

# Test block number
BLOCK_NUM=12911679

# Function to print test results
function print_result {
  if [ $1 -ne 0 ]; then
    echo "❌ $2 failed"
  else
    echo "✅ $2 passed"
  fi
}

# Start the test script
echo "Starting integration tests..."

# 1. Test subscribing to an address
echo "Testing /subscribe endpoint..."
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/subscribe?address=$ADDRESS_1")
if [ "$RESPONSE" -eq 200 ]; then
  print_result 0 "/subscribe for $ADDRESS_1"
else
  print_result 1 "/subscribe for $ADDRESS_1"
fi

# 2. Test subscribing to the same address again (should return conflict)
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/subscribe?address=$ADDRESS_1")
if [ "$RESPONSE" -eq 409 ]; then
  print_result 0 "/subscribe conflict for $ADDRESS_1"
else
  print_result 1 "/subscribe conflict for $ADDRESS_1"
fi

# 3. Test fetching transactions (should be empty at first)
echo "Testing /transactions endpoint..."
TRANSACTIONS=$(curl -s "$BASE_URL/transactions?address=$ADDRESS_1")
if [[ "$TRANSACTIONS" == 'null' ]]; then
  print_result 0 "/transactions for $ADDRESS_1"
else
  print_result 1 "/transactions for $ADDRESS_1"
fi

# 4. Test getting the current block (initially should be 0)
echo "Testing /current-block endpoint..."
CURRENT_BLOCK=$(curl -s "$BASE_URL/current-block" | grep -o '[0-9]*')
if [ "$CURRENT_BLOCK" -eq 0 ]; then
  print_result 0 "/current-block"
else
  print_result 1 "/current-block"
fi

# 5. Test parsing a block
echo "Testing /parse-block endpoint..."
RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/parse-block?block=$BLOCK_NUM")
if [ "$RESPONSE" -eq 200 ]; then
  print_result 0 "/parse-block for block $BLOCK_NUM"
else
  print_result 1 "/parse-block for block $BLOCK_NUM"
fi

# 6. Fetch transactions again to ensure block parsing worked
echo "Testing /transactions after block parse..."
TRANSACTIONS=$(curl -s "$BASE_URL/transactions?address=$ADDRESS_1")
if [[ "$TRANSACTIONS" != "[]" ]]; then
  print_result 0 "/transactions for $ADDRESS_1 after block parsing"
else
  print_result 1 "/transactions for $ADDRESS_1 after block parsing"
fi

# Test finished
echo "Integration tests completed!"



