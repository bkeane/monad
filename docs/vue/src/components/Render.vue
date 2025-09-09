<script setup>
import { CodeBlock, NewThought } from 'vue-tufte';
</script>

<template>
    <section id="about">
        <h2>About Templating</h2>
        <p>
            <NewThought>Template rendering</NewThought> provides a means to DRY configuration for
            cross account and cross service usage. At deploy-time almost all configuration is passed 
            through the templating engine.
        </p>
        <p>
            The render commands are provided as a means of discovery and debugging.
        </p>
    </section>

    <section id="list">
        <h2>List</h2>
        <p>
            Show Available Interpolations
        </p>
        <CodeBlock language="bash">
{{`monad render list
╭───────────────────┬────────────────────────────────────────╮
│Template           │Value                                   │
├───────────────────┼────────────────────────────────────────┤
│\{\{.Account.Id\}\}    │775231628891                            │
│\{\{.Account.Region\}\}│us-west-2                               │
│\{\{.Git.Repo\}\}      │monad                                   │
│\{\{.Git.Owner\}\}     │bkeane                                  │
│\{\{.Git.Branch\}\}    │main                                    │
│\{\{.Git.Sha\}\}       │68adc19d466f7da1efbabed7afe5f894af922679│
│\{\{.Service.Name\}\}  │monad                                   │
│\{\{.Resource.Name\}\} │monad-main-monad                        │
│\{\{.Resource.Path\}\} │monad/main/monad                        │
│\{\{.Ecr.Id\}\}        │775231628891                            │
│\{\{.Ecr.Region\}\}    │us-west-2                               │
╰───────────────────┴────────────────────────────────────────╯`}}
        </CodeBlock>
    </section>

    <section id="string">
        <h2>String</h2>
        <p>Render template strings with contextual data:</p>
        <CodeBlock language="bash">
{{`monad render string "\{\{.Git.Owner\}\}/\{\{.Git.Repo\}\}:\{\{.Git.Branch\}\}"`}}
        </CodeBlock>
    </section>

    <section id="file">
        <h2>File</h2>
        <p>Render template files with contextual data:</p>
        <CodeBlock language="bash">
{{ `monad render file policy.json.tmpl` }}
        </CodeBlock>
    </section>
</template>