Describe "EventBridge"
    target="https://${MONAD_HOST}/${MONAD_REPO}/${MONAD_BRANCH}/${MONAD_SERVICE}"
    event_id=$(uuidgen)

    It "Event Sent"
        When call emit_test_event $event_id
        The status should be success
        The output should be present
    End

    It "Event Received"
        When call curl_sigv4_until 200 "$target/function/log_group/tail?n=50&grep=$event_id&expect=true"
        The output should include 200
        The status should be success
    End
End