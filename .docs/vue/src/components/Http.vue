<script setup>
import { CodeBlock, Figure } from 'vue-tufte';
import webserver from '../../assets/diagrams/deployment-webserver-1.png';
</script>

<template>
    <section id="requirements">
        <h2>Requirements</h2>
        <p>An AWS API Gateway V2 HTTP proxy.</p>
    </section>

        <section>
    <Figure :src="webserver"></Figure>
    </section>
    
    <section id="api">
        <h2>Api</h2>
        <p>The presence of the api flag will mount the lambda to a reverse proxy:</p>
        <CodeBlock language="bash">
{{`monad deploy --api $name`}}
        </CodeBlock>
    </section>
    
    <section id="route">
        <h2>Route</h2>
        <p>Providing the route flag allows for setting the proxy path:</p>
        <CodeBlock language="bash">
{{`monad deploy --api $name --route $pattern`}}
        </CodeBlock>
    </section>
    
    <section id="auth">
        <h2>Auth</h2>
        <p>Providing the auth flag allows for setting the authorization type:</p>
        <CodeBlock language="bash">
{{`monad deploy --api $name --auth $type`}}
        </CodeBlock>
    </section>
    
    <section id="convention">
        <h2>Convention</h2>
        
        <h3>Routes</h3>
        <p>The default route pattern uses Git context for automatic path generation:</p>
        <CodeBlock language="bash">{{`"ANY /\{\{.Git.Repo\}\}/\{\{.Git.Branch\}\}/\{\{.Service.Name\}\}/{proxy+}""`}}
        </CodeBlock>
        
        <h3>Headers</h3>
        <p>The path prefix is forwarded downstream for proxy aware routers (etc):</p>
        <CodeBlock language="bash">
{{`X-Forwarded-Prefix: /repo/branch/service`}}
        </CodeBlock>

        <h3>Auth</h3>
        <p>The default auth type is AWS_IAM which validates sigv4 signed requests.</p>
    </section>
    
    <section id="configuration">
        <h2>Configuration</h2>
        
        <h3>Routes & Auth</h3>
        <CodeBlock language="bash">
{{`# Alternate route
monad deploy --api my-gateway --route "ANY /\{\{.Service.Name\}\}/{proxy+}"

# Alternate auth (fully public)
monad deploy --api my-gateway --auth none

# Public and Private routes
monad deploy --api my-gateway \\
    --route "ANY /api/{proxy+}" --auth aws_iam \\
    --route "ANY /public/{proxy+}" --auth none

# Say you have a custom authorizer named bearer
monad deploy --api my-gateway --auth bearer
` }}
        </CodeBlock>
    </section>
</template>