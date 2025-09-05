<script setup>
import { CodeBlock } from 'vue-tufte';
</script>

<template>
    <section id="requirements">
        <h2>Requirements</h2>
        <p>The deployable artifact of monad is an OCI image hosted in ECR.</p>
    </section>

    <section id="init">
        <h2>Init</h2>
    </section>

    <section id="publish">
        <h2>Publish</h2>
        <p>Monad provides a simple means to produce conventional image tags for builds.</p>
        <code-block language="bash">
{{`# Default gist:
tag=$(monad ecr tag)  # {id}.dkr.ecr.{region}.amazonaws.com/{owner}/{repo}/{service}:{branch}
docker build -t $tag . --push`}}
        </code-block>
    </section>
    
    <section id="deploy">
        <h2>Deploy</h2>
        <p>Deploy artifact as an invokable Lambda function:</p>
        <CodeBlock language="bash">
{{`monad deploy`}}
        </CodeBlock>
    </section>
    
    <section id="convention">
        <h2>Convention</h2>
        <h3>Registry</h3>
        <p>The default target registry is set to that of the caller AWS account and caller region.</p>
        <CodeBlock language="bash">
{{`# Default gist:
ecr_id=$(aws sts get-caller-identity --query Account --output text)
ecr_region=$(aws configure get region)
echo "\${ecr_id}.dkr.ecr.\${ecr_region}.amazonaws.com"`}}      
        </CodeBlock>

        <h3>Image</h3>
        <p>The default image path and tag are constructed from git data.</p>
        <CodeBlock language="bash">
{{`# Default gist:
owner=$(git config --get remote.origin.url | cut -d: -f2 | cut -d/ -f1)
repo=$(basename -s .git "$(git config --get remote.origin.url)")
branch=$(git rev-parse --abbrev-ref HEAD)
service=$(basename "$PWD")
echo "\${owner}/\${repo}/\${service}:\${branch}"`}}
        </CodeBlock>
    </section>
    
    <section id="configuration">
        <h2>Configuration</h2>
        
        <h3>Registry</h3>
        <p>Generate publish tag for central repository:</p>
        <CodeBlock language="bash">
{{`monad ecr tag --ecr-id 123456789012 --service myapp`}}
        </CodeBlock>

        <p>Deploy artifact from a central repository:</p>
        <CodeBlock language="bash">
            {{`monad deploy --ecr-id 123456789012 --service myapp`}}
        </CodeBlock>
        
        <h3>Image</h3>
        <p>Mutate image path and tag by overriding git data:</p>
        <CodeBlock language="bash">
{{`monad ecr tag --owner myorg --repo myproject --branch production --service myapp`}}
        </CodeBlock>
        <p>Or provide a full image path:</p>
        <CodeBlock language="bash">
{{`monad deploy --image myorg/myproject/myservice:v1.0.0 --service myapp`}}
        </CodeBlock>
    </section>
</template>