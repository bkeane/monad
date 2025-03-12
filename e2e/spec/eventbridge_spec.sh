Include spec/helpers.sh

Describe "EventBridge"
  branch="$(git rev-parse --abbrev-ref HEAD)"
  sha="$(git rev-parse HEAD)"
  path="monad/${branch}/echo"
  host="prod.kaixo.io"
  event_id=$(uuidgen)

  It "monad deploy --api kaixo --rule file://rule.json.tmpl" --policy file://policy.json.tmple
    When call monad --chdir echo deploy --api kaixo --rule file://rule.json.tmpl --policy file://policy.json.tmpl
    The status should be success
  End

  It "Health"
    When call curl_sigv4_until 200 https://${host}/${path}/health
    The output should eq 200
    The status should be success
  End

  Describe "Event"
    It "Event Sent"
      When call emit_test_event $event_id
      The status should be success
      The output should be present
    End

    It "Event Received"
      When call curl_sigv4_until 200 "https://${host}/${path}/function/log_group/tail?n=50&grep=$event_id&expect=true"
      The output should include 200
      The status should be success
    End
  End
End