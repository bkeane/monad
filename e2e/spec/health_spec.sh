Include helpers.sh

Describe "Health"
  host=$(resolve_api_domain $MONAD_API)

  It "aws_auth"
    When call curl_sigv4_until 200 "https://${host}/monad/${MONAD_BRANCH}/echo/health"
    The output should include "ok"
    The status should be success
  End

  It "oauth_auth"
    When call curl_oauth_until 200 "https://${host}/monad/${MONAD_BRANCH}/echo-oauth/health"
    The output should include "ok"
    The status should be success
  End
End