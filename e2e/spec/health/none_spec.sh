Describe "Health (no auth)"
  target="https://${MONAD_HOST}/${MONAD_REPO}/${MONAD_BRANCH}/${MONAD_SERVICE}/public/health"
  It "${MONAD_SERVICE}"
    When call curl_until 200 "${target}"
    The output should include 200
    The status should be success
  End
End