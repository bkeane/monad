Include spec/helpers.sh

Describe "Auth"
  branch="$(git rev-parse --abbrev-ref HEAD)"
  host="prod.kaixo.io"
  path="monad/${branch}/echo"

  Describe "Authorizers"
    Describe "NONE"
      It "monad deploy --api $API_NAME --auth none"
        When call monad --chdir echo deploy --api $API_NAME --auth none
        The status should be success
      End

      It "should allow unauthorized requests"
        When call curl_retry_status "https://${host}/${path}/"
        The output should eq "200"
        The status should be success
      End
    End

    Describe "AWS_IAM"
      It "monad deploy --api $API_NAME --auth aws_iam"
        When call monad --chdir echo deploy --api $API_NAME --auth aws_iam
        The status should be success
      End

      It "should deny requests without sigv4 signing"
        When call curl_until_failure "https://${host}/${path}/"
        The output should eq "403"
        The status should be failure
      End

      It "should allow requests with sigv4 signing"
        When call curl_retry_sigv4_status "https://${host}/${path}/"
        The output should eq "200"
        The status should be success
      End
    End

    Describe "JWT"
      It "monad deploy --api kaixo --auth auth0"
        When call monad --chdir echo deploy --api kaixo --auth auth0
        The status should be success
      End

      It "should deny requests without bearer token"
        When call curl_until_failure "https://${host}/${path}/"
        The output should eq "403"
        The status should be failure
      End

      It "should allow requests with oauth bearer token"
        When call curl_retry_oauth_status "https://${host}/${path}/"
        The output should eq "200"
        The status should be success
      End
    End
  End
End
