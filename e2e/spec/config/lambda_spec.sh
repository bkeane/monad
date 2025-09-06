Describe "Lambda"
    target="https://${MONAD_HOST}/${MONAD_REPO}/${MONAD_BRANCH}/${MONAD_SERVICE}"
    resource_name="${MONAD_REPO}-${MONAD_BRANCH}-${MONAD_SERVICE}"
    resource_path="${MONAD_REPO}/${MONAD_BRANCH}/${MONAD_SERVICE}"
    image_path="${MONAD_OWNER}/${MONAD_REPO}/${MONAD_SERVICE}"

    It "Name"
        When call curl_sigv4 $target/function/name
        The output should eq "${resource_name}"
        The status should be success
    End

    Describe "IAM"
        It "Role"
            When call curl_sigv4 $target/function/role/name
            The output should eq "${resource_name}"
            The status should be success
        End

        It "Policy"
            When call curl_sigv4 $target/function/role/policies/${resource_name}
            The output should include "${resource_name}"
            The status should be success
        End
    End

    Describe "Resources"
        It "Memory"
            When call curl_sigv4 $target/function/memory
            The output should eq "256"
            The status should be success
        End

        It "Disk"
            When call curl_sigv4 $target/function/disk
            The output should eq "1024"
            The status should be success
        End

        It "Timeout"
            When call curl_sigv4 $target/function/timeout
            The output should eq "10"
            The status should be success
        End

        It "Log Group"
            When call curl_sigv4 $target/function/log_group
            The output should eq "/aws/lambda/${resource_name}"
            The status should be success
        End
    End

    Describe "Env"
        It "MONAD_RESOURCE_NAME"
            When call curl_sigv4 $target/env/MONAD_RESOURCE_NAME 
            The output should eq "${resource_name}"
            The status should be success
        End

        It "MONAD_RESOURCE_PATH"
            When call curl_sigv4 $target/env/MONAD_RESOURCE_PATH
            The output should eq "${resource_path}"
            The status should be success
        End

        It "MONAD_REPO"
            When call curl_sigv4 $target/env/MONAD_REPO
            The output should eq "${MONAD_REPO}"
            The status should be success
        End

        It "MONAD_SHA"
            When call curl_sigv4 $target/env/MONAD_SHA
            The output should eq "${MONAD_SHA}"
            The status should be success
        End

        It "MONAD_BRANCH"
            When call curl_sigv4 $target/env/MONAD_BRANCH  
            The output should eq "${MONAD_BRANCH}"
            The status should be success
        End

        It "MONAD_SERVICE"
            When call curl_sigv4 $target/env/MONAD_SERVICE
            The output should eq "${MONAD_SERVICE}"
            The status should be success
        End
    End

    Describe "Tags"
        It "Monad"
            When call curl_sigv4 $target/function/tags/Monad  
            The output should eq "true"
            The status should be success
        End
        
        It "Service"
            When call curl_sigv4 $target/function/tags/Service  
            The output should eq "${MONAD_SERVICE}"
            The status should be success
        End

        It "Owner"
            When call curl_sigv4 $target/function/tags/Owner  
            The output should eq "bkeane"
            The status should be success
        End

        It "Repo"
            When call curl_sigv4 $target/function/tags/Repo  
            The output should eq "${MONAD_REPO}"
            The status should be success
        End

        It "Branch"
            When call curl_sigv4 $target/function/tags/Branch  
            The output should eq "${MONAD_BRANCH}"
            The status should be success
        End

        It "Sha"
            When call curl_sigv4 $target/function/tags/Sha  
            The output should eq "${MONAD_SHA}"
            The status should be success
        End
    End
End