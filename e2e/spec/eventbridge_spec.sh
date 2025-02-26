Include spec/helpers.sh

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

Describe "EventBridge"
  branch="$(git rev-parse --abbrev-ref HEAD)"
  sha="$(git rev-parse HEAD)"
  path="monad/${branch}/echo"
  host="prod.kaixo.io"
  event_id=$(uuidgen)

  It "monad deploy --api kaixo --rule file://rule.json" --policy file://policy.json
    When call monad --chdir echo deploy --api kaixo --rule file://rule.json --policy file://policy.json
    The status should be success
  End

  Describe "Event"
    It "Event Sent"
      When call emit_test_event $event_id
      The status should be success
      The output should be present
    End

    It "Event Received"
      When call curl_retry_sigv4_status "https://${host}/${path}/function/log_group/tail?n=50&grep=$event_id&expect=true"
      The output should include "200"
      The status should be success
    End
  End
End