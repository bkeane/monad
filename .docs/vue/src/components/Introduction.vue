<script setup>
import { CodeBlock } from 'vue-tufte';
</script>

<template>
    <section id="installation">
        <h2>Installation</h2>
        
        <h3>GitHub Releases</h3>
        <p>Download the latest binary from <a href="https://github.com/bkeane/monad/releases" target="_blank">GitHub releases</a>.</p>
        
        <h3>Homebrew</h3>
        <CodeBlock language="bash">
{{`brew tap bkeane/tools
brew install bkeane/tools/monad`}}
        </CodeBlock>

        <h3>GitHub Actions</h3>
        <CodeBlock language="yaml">
{{`- uses: bkeane/monad-action@v1`}}
        </CodeBlock>
    </section>

    <section id="usage">
        <h2>Usage</h2>
        
        <CodeBlock language="bash">
{{`monad --help`}}
        </CodeBlock>

        <h2 id="convention">Convention</h2>
        <p>All configurations have conventional defaults, for example:</p>
        
        <ul>
            <li>Service name defaults to directory name</li>
            <li>Git info (owner/repo/branch/SHA) auto-detected from repository</li>
            <li>AWS settings derived from your SDK configuration</li>
            <li>Conservative resource limits (128MB memory, 3s timeout)</li>
        </ul>
        
        <h2 id="configuration">Configuration</h2>
        <p>Every convention can be overridden using <strong>command-line flags</strong> or <strong>environment variables</strong>. All environment variables follow the <code>MONAD_*</code> pattern. Use <code>monad help [command]</code> to see all available options.</p>
    </section>
</template>