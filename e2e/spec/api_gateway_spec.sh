Include spec/helpers.sh

Describe "Auth"
  branch="$(git rev-parse --abbrev-ref HEAD)"
  host="prod.kaixo.io"
  path="monad/${branch}/echo"

  Describe "Authorizers"
    Describe "NONE"
      It "monad --chdir echo deploy --api $API_NAME --auth none"
        When call monad --chdir echo deploy --api $API_NAME --auth none
        The status should be success
      End

      It "should allow unauthorized requests"
        When call curl_until 200 "https://${host}/${path}/"
        The output should eq 200
        The status should be success
      End
    End

    Describe "AWS_IAM"
      It "monad --chdir echo deploy --api $API_NAME --auth aws_iam"
        When call monad --chdir echo deploy --api $API_NAME --auth aws_iam
        The status should be success
      End

      It "should deny requests without sigv4 signing"
        When call curl_until 403 "https://${host}/${path}/"
        The output should eq 403
        The status should be success
      End

      It "should allow requests with sigv4 signing"
        When call curl_sigv4_until 200 "https://${host}/${path}/"
        The output should eq 200
        The status should be success
      End
    End

    Describe "JWT"
      It "monad --chdir echo deploy --api kaixo --auth auth0"
        When call monad --chdir echo deploy --api kaixo --auth auth0
        The status should be success
      End

      It "should deny requests without bearer token"
        When call curl_until 401 "https://${host}/${path}/" 
        The output should eq 401
        The status should be success
      End

      It "should allow requests with oauth bearer token"
        When call curl_oauth_until 200 "https://${host}/${path}/" 
        The output should eq 200
        The status should be success
      End
    End
  End

  Describe "Headers"
    It "monad --chdir echo deploy --api $API_NAME"
      When call monad --chdir echo deploy --api $API_NAME
      The status should be success
    End

    It "Health"
      When call curl_sigv4_until 200 "https://${host}/${path}/health"
      The output should eq 200
      The status should be success
    End

    It "X-Forwarded-Prefix"
      When call curl_sigv4 "https://${host}/${path}/headers/x-forwarded-prefix"  
      The output should include "$path"
      The status should be success
    End
  End
End
