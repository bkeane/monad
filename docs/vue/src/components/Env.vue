<script setup>
import { CodeBlock, NewThought, Sidenote } from 'vue-tufte';
</script>

<template>
    <section>
        <h2>Environment Variables</h2>
        <p>
        <NewThought>Environment configuration</NewThought> allows you to pass runtime configuration 
        to your Lambda functions without hardcoding values in your application code.
        </p>
    </section>

    <section id="dotenv">
        <h2>Dotenv files</h2>
        <p>By default every deployment has the following dotenv file exported:</p>
        <CodeBlock language="dotenv">
{{`GIT_REPO=\{\{.Git.Repo\}\}
GIT_BRANCH=\{\{.Git.Branch\}\}
GIT_SHA=\{\{.Git.Sha\}\}
SERVICE_NAME=\{\{.Service.Name\}\}`}}
        </CodeBlock>
        <p>To provide a custom dotenv file:</p>
        <CodeBlock language="bash">
{{`monad deploy --env .env.tmpl`}}
        </CodeBlock>
    </section>

    <section id="Entrypoint">
        <h2>Entrypoint method</h2>
        <p>
            Dotenv files have limitations that become self evident with use.
            Not least of which being that they are not appropriate for secret material.
        </p>
        <p>
            A viable and complimentary approach is to move environment configuration concerns immediately
            before runtime of your application.
        </p>
        <CodeBlock language="docker">
{{`COPY --from=segment/chamber:2 /chamber /bin/chamber
ENTRYPOINT ["/bin/chamber", "exec", "service", "--"]
CMD ["python", "main.py"]`}}
        </CodeBlock>
        <p>
            Here we see the use of Chamber<Sidenote><a href="https://github.com/segmentio/chamber?tab=readme-ov-file">chamber docs</a></Sidenote>
            as a container execution manager. In this configuration it is exporting variables under the `service` namespace
            in AWS ssm parameter store<Sidenote><a href="https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html">parameter store docs</a></Sidenote>.
        </p>
        <p>
            Chamber is a particularly good choice as it provides an operational interface to manage secrets/configuration
            directly from the cli.
        </p>
        <p>
            There are other tools and approaches depending on needs, but this is an exercise for the reader as there
            is no one-size-fits-all solution.
        </p>
    </section>
</template>