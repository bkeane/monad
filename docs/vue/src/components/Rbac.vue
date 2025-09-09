<script setup>
import { CodeBlock, NewThought, Sidenote } from 'vue-tufte';
</script>

<template>
    <section id="about">
        <h2>About RBAC</h2>
        <p>
            <NewThought>AWS IAM roles</NewThought> are both required at deploy-time and automatically surfaced at run-time
            in AWS Lambdas. This provides a powerful mechanism to safely and precisely permission your code for AWS service
            access. It also provides a low effort and secure mechanism for machine-to-machine authenticated calling
            via AWS sigv4 signing<Sidenote><a href="https://docs.aws.amazon.com/AmazonS3/latest/API/sig-v4-authenticating-requests.html">sigv4 docs</a></Sidenote>.  
        </p>
    </section>
    <section id="role">
        <h2>Role</h2>
        <p>The default role<Sidenote><a href="https://docs.aws.amazon.com/lambda/latest/dg/lambda-intro-execution-role.html">Lambda execution role</a></Sidenote> allows Lambda service to assume it.</p>
        <p>Apply a custom IAM role trust policy:</p>
        <CodeBlock language="bash">
{{`monad deploy --role ./role.json.tmpl`}}
        </CodeBlock>
    </section>
    
    <section id="policy">
        <h2>Policy</h2>
        <p>The default policy<Sidenote><a href="https://docs.aws.amazon.com/lambda/latest/dg/lambda-permissions.html">Lambda policy</a></Sidenote> grants full CloudWatch Logs access<Sidenote><a href="https://docs.aws.amazon.com/service-authorization/latest/reference/reference_policies_actions-resources-contextkeys.html">AWS actions & conditions</a></Sidenote> for the function's log group.</p>
        <p>Apply a custom IAM policy template:</p>
        <CodeBlock language="bash">
{{`monad deploy --policy ./policy.json.tmpl`}}
        </CodeBlock>
    </section>
    
    <section id="boundary">
        <h2>Boundary</h2>
        <p>No permissions boundary<Sidenote><a href="https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_boundaries.html">Boundary policies</a></Sidenote> is applied by default.</p>
        <p>Apply an existant permissions boundary for role:</p>
        <CodeBlock language="bash">
{{`monad deploy --boundary BoundaryPolicyName`}}
        </CodeBlock>
    </section>
</template>