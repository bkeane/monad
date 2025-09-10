<script setup>
import { CodeBlock, Figure } from 'vue-tufte';
import webserver from '../../assets/diagrams/deployment-http-1.png';
</script>

<template>
    <section id="about">
        <h2>About HTTP</h2>
        <p>
            <NewThought>The Api Gateway v2 HTTP Proxy</NewThought> service is used by Monad to provide
            both HTTP reverse proxy functionality, as well as infrastructure level authorization.
        </p>
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

    <section id="auth">
        <h2>Auth</h2>
        <p>By default the auth type is AWS_IAM
            <Sidenote><a href="https://how.wtf/aws-sigv4-requests-with-curl.html">curl sigv4</a></Sidenote>
            <Sidenote><a href="https://github.com/legal90/awscurl">awscurl</a></Sidenote>.</p>
        <p>Providing the auth flag allows for setting the authorization<Sidenote><a href="https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-access-control.html">
            api gateway authorizers
        </a></Sidenote> type:</p>
        <CodeBlock language="bash">
{{`monad deploy --api $name --auth $type`}}
        </CodeBlock>

        <h3>Examples</h3>
        <CodeBlock language="bash">
{{`# Fully public
monad deploy --api $name --auth none

# IAM Authenticated
monad deploy --api $name --auth aws_iam

# Custom authorizer
# given you have an authorizer named "bearer"
monad deploy --api my-gateway --auth bearer 
`}}
        </CodeBlock>
    </section>
    
    <section id="route">
        <h2>Route</h2>
        <p>The default route pattern<Sidenote>
            <a href="https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-develop-routes.html">greedy proxy route keys</a>
        </Sidenote> uses Git context ensuring deployments do not conflict:</p>
        <CodeBlock language="bash">{{`monad render string "ANY /\{\{.Git.Repo\}\}/\{\{.Git.Branch\}\}/\{\{.Service.Name\}\}/{proxy+}"`}}
        </CodeBlock>
        <p>Providing the route flag allows for setting the proxy path:</p>
        <CodeBlock language="bash">
{{`monad deploy --api $name --route $pattern`}}
        </CodeBlock>
        <h3>Examples</h3>
        <CodeBlock language="bash">
{{`# Alternate route
monad deploy --api my-gateway --route /\{\{.Service.Name}\}\}/\{\{.Git.Branch\}\}/{proxy+}

# Private/Public split routing
monad deploy --api my-gateway \\
    --route "ANY /\{\{.Git.Branch\}\}/private/{proxy+}" --auth aws_iam \\
    --route "ANY /\{\{.Git.Branch\}\}/public/{proxy+}" --auth none
` }}
        </CodeBlock>
    </section>

    <section id="header">
        <h2>Header</h2>
        <p>The path prefix header<Sidenote><a href="https://learn.microsoft.com/en-us/aspnet/core/fundamentals/servers/yarp/transforms?view=aspnetcore-9.0#defaults">de facto standard</a></Sidenote> is forwarded downstream for proxy aware routers (etc):</p>
        <CodeBlock language="bash">
{{`# Given you are using the default --route
monad render string "X-Forwarded-Prefix: /\{\{\.Git.Repo\}\}/\{\{.Git.Branch\}\}/\{\{.Service.Name\}\}"`}}
        </CodeBlock>
    </section>
</template>