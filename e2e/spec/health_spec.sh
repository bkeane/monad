Describe "Health"
  It "echo"
    When call curl_sigv4_until 200 "https://${MONAD_HOST}/monad/${MONAD_BRANCH}/echo/health"
    The output should include 200
    The status should be success
  End

  It "echo-oauth"
    When call curl_oauth_until 200 "https://${MONAD_HOST}/monad/${MONAD_BRANCH}/echo-oauth/health"
    The output should include 200
    The status should be success
  End

  It "echo-vpc"
    When call curl_sigv4_until 200 "https://${MONAD_HOST}/monad/${MONAD_BRANCH}/echo-vpc/health"
    The output should include 200
    The status should be success
  End
End