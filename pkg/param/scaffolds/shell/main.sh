#!/bin/sh

# https://docs.aws.amazon.com/lambda/latest/dg/runtimes-walkthrough.html

set -eu

export AWS_PAGER=""
export AWS_DEFAULT_OUTPUT=json

handler() {
  entrypoint=$(echo "$EVENT_DATA" | jq -r '.entrypoint')
  cmd=$(echo "$EVENT_DATA" | jq -r '.cmd')
  
  stdout_file=$(mktemp)
  stderr_file=$(mktemp)
  
  eval "$entrypoint $cmd" > "$stdout_file" 2> "$stderr_file"
  
  exit_code=$?
  stdout=$(cat "$stdout_file")
  stderr=$(cat "$stderr_file")
  response=$(jq -n --arg out "$stdout" --arg err "$stderr" --arg exit_code "$exit_code" '{stdout: $out, stderr: $err, exit_code: $exit_code}')
  
  rm "$stdout_file" "$stderr_file"
  
  echo $response
}

while true
do
  HEADERS="$(mktemp)"
  EVENT_DATA=$(curl -sS -LD "$HEADERS" "http://${AWS_LAMBDA_RUNTIME_API}/2018-06-01/runtime/invocation/next")
  REQUEST_ID=$(grep -Fi Lambda-Runtime-Aws-Request-Id "$HEADERS" | tr -d '[:space:]' | cut -d: -f2)
  RESPONSE=$(handler $EVENT_DATA)
  curl "http://${AWS_LAMBDA_RUNTIME_API}/2018-06-01/runtime/invocation/$REQUEST_ID/response" -d "$RESPONSE"
done