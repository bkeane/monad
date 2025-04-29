# Describe "Configuration"
#   target="https://${MONAD_HOST}/monad/${MONAD_BRANCH}/echo"

#   Describe "Lambda"
#     It "Name"
#       When call curl_sigv4 $target/function/name
#       The output should eq "monad-${MONAD_BRANCH}-echo"
#       The status should be success
#     End

#     Describe "IAM"
#       It "Role"
#         When call curl_sigv4 $target/function/role/name
#         The output should eq "monad-${MONAD_BRANCH}-echo"
#         The status should be success
#       End

#       It "Policy"
#         When call curl_sigv4 $target/function/role/policies/monad-${MONAD_BRANCH}-echo
#         The output should include "monad-${MONAD_BRANCH}-echo"
#         The status should be success
#       End
#     End

#     Describe "Resources"
#       It "Memory"
#         When call curl_sigv4 $target/function/memory
#         The output should eq "256"
#         The status should be success
#       End

#       It "Timeout"
#         When call curl_sigv4 $target/function/timeout
#         The output should eq "10"
#         The status should be success
#       End

#       It "Disk"
#         When call curl_sigv4 $target/function/disk
#         The output should eq "1024"
#         The status should be success
#       End

#       It "Log Group"
#         When call curl_sigv4 $target/function/log_group
#         The output should eq "/aws/lambda/monad/${MONAD_BRANCH}/echo"
#         The status should be success
#       End
#     End

#     Describe "Env"
#       It "MONAD_NAME"
#         When call curl_sigv4 $target/env/MONAD_NAME  
#         The output should eq "monad-${MONAD_BRANCH}-echo"
#         The status should be success
#       End

#       It "MONAD_PATH"
#         When call curl_sigv4 $target/env/MONAD_PATH  
#         The output should eq "monad/${MONAD_BRANCH}/echo"
#         The status should be success
#       End

#       It "MONAD_BRANCH"
#         When call curl_sigv4 $target/env/MONAD_BRANCH  
#         The output should eq "${MONAD_BRANCH}"
#         The status should be success
#       End
      
#       It "MONAD_SHA"
#         When call curl_sigv4 $target/env/MONAD_SHA  
#         The output should eq "${MONAD_SHA}"
#         The status should be success
#       End

#       It "MONAD_CUSTOM"
#         When call curl_sigv4 $target/env/MONAD_CUSTOM  
#         The output should eq "present"
#         The status should be success
#       End
#     End

#     Describe "Tags"
#       It "Owner"
#         When call curl_sigv4 $target/function/tags/Owner  
#         The output should eq "bkeane"
#         The status should be success
#       End

#       It "Repository"
#         When call curl_sigv4 $target/function/tags/Repository  
#         The output should eq "monad"
#         The status should be success
#       End
      
#       It "Branch"
#         When call curl_sigv4 $target/function/tags/Branch  
#         The output should eq "${MONAD_BRANCH}"
#         The status should be success
#       End

#       It "Service"
#         When call curl_sigv4 $target/function/tags/Service  
#         The output should eq "echo"
#         The status should be success
#       End
      
#       It "Sha"
#         When call curl_sigv4 $target/function/tags/Sha  
#         The output should eq "${MONAD_SHA}"
#         The status should be success
#       End
#     End
#   End

#   Describe "API Gateway"
#     Describe "Headers"
#       It "X-Forwarded-Prefix"
#         When call curl_sigv4 $target/headers/x-forwarded-prefix
#         The output should include "$path"
#         The status should be success
#       End
#     End
#   End

#   Describe "EventBridge"
#     event_id=$(uuidgen)

#     It "Event Sent"
#       When call emit_test_event $event_id
#       The status should be success
#       The output should be present
#     End

#     It "Event Received"
#       When call curl_sigv4_until 200 "$target/function/log_group/tail?n=50&grep=$event_id&expect=true"
#       The output should include 200
#       The status should be success
#     End
#   End
# End