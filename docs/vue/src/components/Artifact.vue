<script setup>
import { CodeBlock, NewThought, Sidenote, Figure } from 'vue-tufte';
import basic from '../../assets/diagrams/deployment-basic-1.png'
</script>

<template>
    <section id="about">
        <h2>About Artifacts</h2>
        <p>
        <NewThought>OCI Images stored on ECR</NewThought> are the deployable artifact of Monad. 
        If convention is followed these artifacts have an entrypoint of a webserver. So
        even though Monad is a serverless deployment approach, its standard artifact is compatible
        with any container runtime capable of managing containerized webservers<Sidenote>
                <a href="https://github.com/awslabs/aws-lambda-web-adapter">AWS Lambda Web Adapter</a>
        </Sidenote>.
        </p>
    </section>

    <section id="init">
        <h2>Init</h2>
        <p>Scaffold a new monad project from a language template:</p>
        <CodeBlock language="bash">
{{`monad init $language
.
├── Dockerfile
└── main.ext

1 directory, 2 files`}}
        </CodeBlock>
        <h3>Templates</h3>
        <p>To write out one or many of the default configuration templates for modification:</p>
        <CodeBlock language="bash">
{{`monad init $language --env --role --policy --rule
.
├── Dockerfile
├── main.ext
├── .env.tmpl
├── policy.json.tmpl
├── role.json.tmpl
└── rule.json.tmpl

1 directory, 5 files`}}
        </CodeBlock>
    </section>

    <section id="publish">
        <h2>Publish</h2>
        <p>Monad provides a simple means to produce conventional image tags for builds.</p>
        <CodeBlock language="bash">
{{`docker build -t $(monad ecr tag) . --push`}}
        </CodeBlock>

        <h3>Registry</h3>
        <p>You can point your artifact for publishing to an alternate repository:</p>
        <CodeBlock language="bash">
{{`monad ecr tag --ecr-id 123456789012`}}
        </CodeBlock>
        <p>The default target registry is set to that of the caller AWS account and caller region.</p>
        <CodeBlock language="bash">
{{`monad render string "\{\{.Ecr.Id\}\}.dkr.ecr.\{\{.Ecr.Region\}\}"`}}      
        </CodeBlock>

        <h3>Image</h3>
        <p>The default image path and tag are constructed from git data and service name.</p>
        <CodeBlock language="bash">
{{`monad render string "\{\{.Git.Owner\}\}/\{\{.Git.Repo\}\}/\{\{.Service.Name\}\}:\{\{.Git.Branch\}\}"`}}
        </CodeBlock>

        <p>Mutate image path and tag by overriding git data:</p>
        <CodeBlock language="bash">
{{`monad ecr tag --owner $owner --repo $repo --branch $branch`}}
        </CodeBlock>

        <p>Or provide a full image path, preserving git data:</p>
        <CodeBlock language="bash">
{{`monad ecr tag --image bkeane/cowsay:1.0.0 --service cowsay`}}
        </CodeBlock>
    </section>
    
    <section id="deploy">
        <h2>Deploy</h2>
        <p>Deploy artifact as an invokable Lambda function:</p>
        <CodeBlock language="bash">
{{`monad deploy`}}
        </CodeBlock>

        <section>
                <Figure :src="basic"></Figure>
                        <p>
        It's worth noting that while you can invoke<Sidenote><a href="https://docs.aws.amazon.com/cli/latest/reference/lambda/invoke.html">aws lambda invoke</a></Sidenote> a conventional image configuration via the cli,
        you will have to pack your request into an appropriate AWS http proxy event<Sidenote>
                <a href="https://docs.aws.amazon.com/apigateway/latest/developerguide/set-up-lambda-proxy-integrations.html#api-gateway-simple-proxy-for-lambda-input-format">aws http poxy event schema</a>
        </Sidenote> for consumption by the AWS Lambda Web Adapter.
        </p>
        </section>
        


        <p>
        It is recommended you simply add an Api Gateway proxy route<Sidenote>
                <a href="/monad/#/http/about">see http section</a>
        </Sidenote> or subscribe it to an event bus<Sidenote>
                <a href="/monad/#/event/about">see event section</a>
        </Sidenote>.
        </p>

        <h3>Registry</h3>
        <p>You can deploy an artifact from an alternate registry:</p>
        <CodeBlock language="bash">
{{`monad deploy --ecr-id 123456789012`}}
        </CodeBlock>

        <h3>Image</h3>
        <p>Mutate image path and tag by overriding git data:</p>
        <CodeBlock language="bash">
{{`monad deploy --owner $owner --repo $repo --branch $branch`}}
        </CodeBlock>

        <p>Or provide a full image path, preserving git data:</p>
        <CodeBlock language="bash">
{{`monad deploy --image bkeane/cowsay:1.0.0 --service cowsay`}}
        </CodeBlock>

        <h3>More</h3>
        <p>If you wish to deploy your artifact to a VPC subnet or adjust provisioned resources see:</p>
        <CodeBlock language="bash">{{ `monad deploy -h` }}</CodeBlock>   
    </section>
</template>