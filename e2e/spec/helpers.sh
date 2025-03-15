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

ensure_aws_creds() {
  if [ -z "$AWS_ACCESS_KEY_ID" ] || [ -z "$AWS_SECRET_ACCESS_KEY" ] || [ -z "$AWS_SESSION_TOKEN" ]; then
    session=$(aws sts get-session-token --duration-seconds 3600)
    export AWS_ACCESS_KEY_ID=$(echo $session | jq -r .Credentials.AccessKeyId)
    export AWS_SECRET_ACCESS_KEY=$(echo $session | jq -r .Credentials.SecretAccessKey)
    export AWS_SESSION_TOKEN=$(echo $session | jq -r .Credentials.SessionToken)
  fi
}

curl_sigv4() {
  ensure_aws_creds
  curl -s --user "$AWS_ACCESS_KEY_ID:$AWS_SECRET_ACCESS_KEY" --aws-sigv4 "aws:amz:us-west-2:execute-api" --header "X-Amz-Security-Token: $AWS_SESSION_TOKEN" "$@"
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
    debug_log=""
    
    debug_log="${debug_log}Looking for API with name '${api_name}'\n"
    api_id=$(aws apigatewayv2 get-apis --query "Items[?Name=='${api_name}'].ApiId" --output text)
    debug_log="${debug_log}Found API ID: '${api_id}'\n"
    
    if [ -z "$api_id" ] || [ "$api_id" = "None" ]; then
        echo -e "$debug_log" >&2
        echo "Error: No API found with name '${api_name}'" >&2
        return 1
    fi

    domains=$(aws apigatewayv2 get-domain-names --query "Items[].DomainName" --output text)
    debug_log="${debug_log}Found domains: ${domains}\n"

    found_domain=""
    
    for domain in $domains; do
        debug_log="${debug_log}Checking mappings for domain: ${domain}\n"
        
        mappings=$(aws apigatewayv2 get-api-mappings \
            --domain-name "${domain}" \
            --query "Items[?ApiId=='${api_id}'].[ApiId,Stage]" \
            --output text)
        
        if [ -n "$mappings" ] && [ "$mappings" != "None" ]; then
            if [ -z "$found_domain" ]; then
                found_domain="$domain"
            fi
            debug_log="${debug_log}Found mapping: ${domain} -> ${api_name}:${api_id}\n"
        fi
    done

    if [ -z "$found_domain" ]; then
        echo -e "$debug_log" >&2
        echo "Error: No custom domain found for API '${api_name}'" >&2
        return 1
    fi

    echo "$found_domain"
    return 0
}