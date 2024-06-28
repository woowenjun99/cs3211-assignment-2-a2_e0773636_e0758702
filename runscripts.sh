#!/bin/bash

if [ "$#" -ne 2 ]; then
  echo "Error: Exactly two arguments are required."
  echo "Usage: $0 <number> <boolean>"
  exit 1
fi

# Check if the first argument is a number
re='^[0-9]+$'
if ! [[ $1 =~ $re ]]; then
  echo "Error: The first argument is not a number."
  exit 1
fi

# Check if the second argument is a boolean
second_arg_lower=$(echo "$2" | tr '[:upper:]' '[:lower:]')
if [ "$second_arg_lower" != "true" ] && [ "$second_arg_lower" != "false" ]; then
  echo "Error: The second argument is not a boolean (true or false)."
  exit 1
fi

if [ "$second_arg_lower" == "true" ]; then
  echo "Generating script..."
  python3 scripts/script.py
fi


YELLOW='\033[0;33m'
RESET='\033[0m'

for ((i=1; i<=$1; i++)); do
  echo
  echo "Iteration: $i"
  echo -e "\n${YELLOW}TEST${RESET} $test_file"
  output=$(./grader ./engine < scripts/generated.in)
done
