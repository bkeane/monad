Describe "API Gateway"
    target="https://${MONAD_HOST}/${MONAD_REPO}/${MONAD_BRANCH}/${MONAD_SERVICE}"
    
    Describe "Headers"
        It "X-Forwarded-Prefix"
            When call curl_sigv4 $target/headers/x-forwarded-prefix
            The output should include "$path"
            The status should be success
        End
    End
End