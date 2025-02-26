export LOG_LEVEL=warn
export API_NAME=kaixo

fetch_client_creds() {
  echo $(aws ssm get-parameter --name "/monad/e2e/auth0" --query "Parameter.Value" --with-decryption --output json | jq -r)
}

fetch_bearer_token() {
  creds="$(fetch_client_creds)"
  client_id=$(echo $creds | jq -r .client_id)
  client_secret=$(echo $creds | jq -r .client_secret)
  endpoint=https://kaixo.auth0.com/oauth/token
  audience=https://kaixo.io

  response=$(curl -s --request POST $endpoint \
    --header 'content-type: application/x-www-form-urlencoded' \
    --data grant_type=client_credentials \
    --data client_id=$client_id \
    --data client_secret=$client_secret \
    --data audience=$audience)

  echo $response | jq -r .access_token
}

curl_retry() {
  curl -s --retry-all-errors --retry 5 "$@"
}

curl_retry_sigv4() {
  session=$(aws sts get-session-token --duration-seconds 3600)
  key=$(echo $session | jq -r .Credentials.AccessKeyId)
  secret=$(echo $session | jq -r .Credentials.SecretAccessKey)
  token=$(echo $session | jq -r .Credentials.SessionToken)
  curl_retry --user "$key:$secret" --aws-sigv4 "aws:amz:us-west-2:execute-api" --header "X-Amz-Security-Token: $token" "$@"
}

curl_retry_oauth() {
  curl_retry --header "Authorization: Bearer $(fetch_bearer_token)" "$@"
}

# For testing auth swapping
curl_retry_status() {
  curl -s --fail --retry-all-errors --retry 7 --retry-delay 2 -o /dev/null -w "%{http_code}" "$@"
}

curl_retry_oauth_status() {
  curl_retry_status --header "Authorization: Bearer $(fetch_bearer_token)" "$@"
}

curl_retry_sigv4_status() {
  session=$(aws sts get-session-token --duration-seconds 3600)
  key=$(echo $session | jq -r .Credentials.AccessKeyId)
  secret=$(echo $session | jq -r .Credentials.SecretAccessKey)
  token=$(echo $session | jq -r .Credentials.SessionToken)
  curl_retry_status --fail --user "$key:$secret" --aws-sigv4 "aws:amz:us-west-2:execute-api" --header "X-Amz-Security-Token: $token" "$@"
}

curl_until_failure() {
  local max_attempts=7
  local attempt=1
  local delay=2
  
  while [ $attempt -le $max_attempts ]; do
    response=$(curl -s -o /dev/null -w "%{http_code}" "$@")
    if [ "$response" != "200" ]; then
      echo $response
      return 1
    fi
    sleep $delay
    attempt=$((attempt + 1))
  done
  
  echo "Failed to get 403/401"
  return 1
}

emit_test_event() {
  aws events put-events --entries '[
    {
      "Source": "shellspec",
      "DetailType": "TestEvent",
      "Detail": "{\"Hello\": \"World\"}",
      "EventBusName": "default"
    }
  ]'
}