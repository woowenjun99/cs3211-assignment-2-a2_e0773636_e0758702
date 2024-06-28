#!/bin/bash

YELLOW='\033[0;33m'
RESET='\033[0m'

for ((i=1; i<=1; i++)); do
  echo "Iteration: $i"
  for test_file in ./tests/*; do
    if [ -f "$test_file" ]; then
      echo -e "\n${YELLOW}TEST${RESET} $test_file"
      output=$(./grader ./engine < "$test_file")
    fi
  done
done
