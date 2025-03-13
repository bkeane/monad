# Include spec/helpers.sh

# Describe "Lambda"
#   branch="$(git rev-parse --abbrev-ref HEAD)"
#   sha="$(git rev-parse HEAD)"
#   path="monad/${branch}/echo"
#   host="prod.kaixo.io"
#   account_id="$(aws sts get-caller-identity --query "Account" --output text)"

#   It "monad deploy --api kaixo"
#     When call monad --chdir echo deploy --disk 1024 --memory 256 --timeout 10 --api kaixo --policy file://policy.json.tmpl --env file://.env.tmpl
#     The status should be success
#   End

#   It "Health"
#     When call curl_sigv4_until 200 https://${host}/${path}/health
#     The output should eq 200
#     The status should be success
#   End

#   It "Name"
#     When call curl_sigv4 https://${host}/${path}/function/name
#     The output should eq "monad-${branch}-echo"
#     The status should be success
#   End

#   Describe "IAM"
#     It "Role"
#       When call curl_sigv4 https://${host}/${path}/function/role/name
#       The output should eq "monad-${branch}-echo"
#       The status should be success
#     End

#     It "Policy"
#       When call curl_sigv4 https://${host}/${path}/function/role/policies/monad-${branch}-echo
#       The output should include "monad-${branch}-echo"
#       The status should be success
#     End
#   End

#   Describe "Resources"
#     It "Memory"
#       When call curl_sigv4 https://${host}/${path}/function/memory
#       The output should eq "256"
#       The status should be success
#     End

#     It "Timeout"
#       When call curl_sigv4 https://${host}/${path}/function/timeout
#       The output should eq "10"
#       The status should be success
#     End

#     It "Disk"
#       When call curl_sigv4 https://${host}/${path}/function/disk
#       The output should eq "1024"
#       The status should be success
#     End

#     It "Log Group"
#       When call curl_sigv4 https://${host}/${path}/function/log_group
#       The output should eq "/aws/lambda/monad/${branch}/echo"
#       The status should be success
#     End
#   End

#   Describe "Env"
#     It "MONAD_NAME"
#       When call curl_sigv4 https://${host}/${path}/env/MONAD_NAME  
#       The output should eq "monad-${branch}-echo"
#       The status should be success
#     End

#     It "MONAD_PATH"
#       When call curl_sigv4 https://${host}/${path}/env/MONAD_PATH  
#       The output should eq "${path}"
#       The status should be success
#     End

#     It "MONAD_BRANCH"
#       When call curl_sigv4 https://${host}/${path}/env/MONAD_BRANCH  
#       The output should eq "${branch}"
#       The status should be success
#     End
    
#     It "MONAD_SHA"
#       When call curl_sigv4 https://${host}/${path}/env/MONAD_SHA  
#       The output should eq "${sha}"
#       The status should be success
#     End

#     It "MONAD_CUSTOM"
#       When call curl_sigv4 https://${host}/${path}/env/MONAD_CUSTOM  
#       The output should eq "present"
#       The status should be success
#     End
#   End

#   Describe "Tags"
#     It "Owner"
#       When call curl_sigv4 https://${host}/${path}/function/tags/Owner  
#       The output should eq "bkeane"
#       The status should be success
#     End

#     It "Repository"
#       When call curl_sigv4 https://${host}/${path}/function/tags/Repository  
#       The output should eq "monad"
#       The status should be success
#     End
    
#     It "Branch"
#       When call curl_sigv4 https://${host}/${path}/function/tags/Branch  
#       The output should eq "${branch}"
#       The status should be success
#     End

#     It "Service"
#       When call curl_sigv4 https://${host}/${path}/function/tags/Service  
#       The output should eq "echo"
#       The status should be success
#     End
    
#     It "Sha"
#       When call curl_sigv4 https://${host}/${path}/function/tags/Sha  
#       The output should eq "${sha}"
#       The status should be success
#     End
#   End
# End