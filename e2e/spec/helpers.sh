export LOG_LEVEL=warn
export MONAD_API=${MONAD_API:-kaixo}
export MONAD_BRANCH=${MONAD_BRANCH:-$(git rev-parse --abbrev-ref HEAD)}
export MONAD_SHA=${MONAD_SHA:-$(git rev-parse HEAD)}


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

curl_oauth() {
  curl -s --header "Authorization: Bearer $(fetch_bearer_token)" "$@"
}

curl_sigv4() {
  session=$(aws sts get-session-token --duration-seconds 3600)
  key=$(echo $session | jq -r .Credentials.AccessKeyId)
  secret=$(echo $session | jq -r .Credentials.SecretAccessKey)
  token=$(echo $session | jq -r .Credentials.SessionToken)
  curl -s --user "$key:$secret" --aws-sigv4 "aws:amz:us-west-2:execute-api" --header "X-Amz-Security-Token: $token" "$@"
}

curl_until() {
  expected_status=$1
  url=$2
  max_tries=${3:-7}  # Default max tries of 7
  delay=${4:-2}  # Default delay of 2 seconds
  
  tries=0
  while [ $tries -lt $max_tries ]; do
    status=$(curl -s -o /dev/null -w "%{http_code}" "$url")
    
    if [ "$status" = "$expected_status" ]; then
      echo "$status"
      return 0
    fi
    
    tries=$((tries + 1))
    
    if [ $tries -lt $max_tries ]; then
      sleep $delay
    fi
  done
  
  echo "$status"
  return 1
}

curl_sigv4_until() {
  expected_status=$1
  url=$2
  max_tries=${3:-7}  # Default max tries of 7
  delay=${4:-2}  # Default delay of 2 seconds
  
  tries=0
  while [ $tries -lt $max_tries ]; do
    status=$(curl_sigv4 -s -o /dev/null -w "%{http_code}" "$url")
    
    if [ "$status" = "$expected_status" ]; then
      echo "$status"
      return 0
    fi
    
    tries=$((tries + 1))
    
    if [ $tries -lt $max_tries ]; then
      sleep $delay
    fi
  done
  
  echo "$status"
  return 1
}

curl_oauth_until() {
  expected_status=$1
  url=$2
  max_tries=${3:-7}  # Default max tries of 7
  delay=${4:-2}  # Default delay of 2 seconds
  
  tries=0
  while [ $tries -lt $max_tries ]; do
    status=$(curl_oauth -s -o /dev/null -w "%{http_code}" "$url")
    
    if [ "$status" = "$expected_status" ]; then
      echo "$status"
      return 0
    fi
    
    tries=$((tries + 1))
    
    if [ $tries -lt $max_tries ]; then
      sleep $delay
    fi
  done
  
  echo "$status"
  return 1
}

emit_test_event() {
  local string=$1
  aws events put-events --entries '[
    {
      "Source": "shellspec",
      "DetailType": "TestEvent", 
      "Detail": "{\"Message\": \"'$string'\"}",
      "EventBusName": "default"
    }
  ]'
}

resolve_api_domain() {
    if [ -z "$1" ]; then
        echo "Error: API name argument is required" >&2
        return 1
    fi

    api_name="$1"
    api_id=""

    echo "DEBUG: Looking for API with name '${api_name}'" >&2
    
    api_id=$(aws apigatewayv2 get-apis --query "Items[?Name=='${api_name}'].ApiId" --output text)
    echo "DEBUG: Found API ID: '${api_id}'" >&2
    
    if [ -z "$api_id" ] || [ "$api_id" = "None" ]; then
        echo "Error: No API found with name '${api_name}'" >&2
        return 1
    fi

    domains=$(aws apigatewayv2 get-domain-names --query "Items[].DomainName" --output text)
    echo "DEBUG: Found domains: ${domains}" >&2

    found_domain=""
    
    for domain in $domains; do
        echo "DEBUG: Checking mappings for domain: ${domain}" >&2
        
        mappings=$(aws apigatewayv2 get-api-mappings \
            --domain-name "${domain}" \
            --query "Items[?ApiId=='${api_id}'].[ApiId,Stage]" \
            --output text)
        
        echo "DEBUG: Mappings result for ${domain}: ${mappings}" >&2
        
        if [ -n "$mappings" ] && [ "$mappings" != "None" ]; then
            if [ -z "$found_domain" ]; then
                found_domain="$domain"
            fi
            echo "DEBUG: Found mapping: ${domain} -> ${api_name}:${api_id}:${stage}" >&2
        fi
    done

    if [ -z "$found_domain" ]; then
        echo "Error: No custom domain found for API '${api_name}'" >&2
        return 1
    fi

    echo "$found_domain"
    return 0
}