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
        <p>Create a new monad project from a template with available language scaffolds:</p>
        <CodeBlock language="bash">
{{`monad init $language`}}
        </CodeBlock>
    </section>

    <section id="publish">
        <h2>Publish</h2>
        <p>Monad provides a simple means to produce conventional image tags for builds.</p>
        <CodeBlock language="bash">
{{`docker build -t $(monad ecr tag) . --push`}}
        </CodeBlock>
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
{{`"$ecr_id.dkr.ecr.$ecr_region.amazonaws.com"`}}      
        </CodeBlock>

        <h3>Image</h3>
        <p>The default image path and tag are constructed from git data and service name.</p>
        <CodeBlock language="bash">
{{`"$owner/$repo/$service:$branch"`}}
        </CodeBlock>
    </section>
    
    <section id="configuration">
        <h2>Configuration</h2>
        
        <h3>Registry</h3>
        <p>Generate publish tag for central repository:</p>
        <CodeBlock language="bash">
{{`monad ecr tag --ecr-id 123456789012`}}
        </CodeBlock>

        <p>Deploy artifact from a central repository:</p>
        <CodeBlock language="bash">
{{`monad deploy --ecr-id 123456789012`}}
        </CodeBlock>
        
        <h3>Image</h3>
        <p>Mutate image path and tag by overriding git data:</p>
        <CodeBlock language="bash">
{{`monad ecr tag --owner $owner --repo $repo --branch $branch`}}
        </CodeBlock>
        <p>Or provide a full image path:</p>
        <CodeBlock language="bash">
{{`monad deploy --image bkeane/cowsay:1.0.0 --service cowsay`}}
        </CodeBlock>
    </section>
</template>