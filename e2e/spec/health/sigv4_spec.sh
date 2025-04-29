Describe "Health (sigv4)"
  It "${MONAD_SERVICE}"
    When call curl_sigv4_until 200 "https://${MONAD_HOST}/${MONAD_REPO}/${MONAD_BRANCH}/${MONAD_SERVICE}/health"
    The output should include 200
    The status should be success
  End
End