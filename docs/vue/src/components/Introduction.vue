<script setup>
import { CodeBlock, Sidenote, Figure, NewThought } from 'vue-tufte';
import monad from '../../assets/diagrams/deployment-complete-1.png'
</script>

<template>
    <section>
        <h2 id="introduction">What is it?</h2>
        <p>
            <NewThought>Monad<Sidenote><a href="https://en.wikipedia.org/wiki/Monadology">monadology</a></Sidenote> is a declarative CLI tool</NewThought> for deploying an OCI image whose entrypoint is a webserver.
        </p>
    </section>

    <section>
        <Figure :src="monad">A Monad deployment</Figure>
    </section>
    
    <section>
        <h3>A Data Pipeline</h3>
        <p>
            Monad can be thought of as a data pipeline feeding an IaC tool.
        </p>
        <p>
            Monad uses two core data sources to drive it's conventions:
            <ol>
                <li>Git context</li>
                <li>AWS caller context<Sidenote><a href="https://docs.aws.amazon.com/sdkref/latest/guide/standardized-credentials.html">aws credential chain</a></Sidenote></li>
            </ol>
            The data from these two sources forms the basis<Sidenote>
                <a href="https://en.wikipedia.org/wiki/Basis_(linear_algebra)">
                    As in maths
                </a>
            </Sidenote> of all resource naming and template data.
        </p>
    </section>
    <section>
        <h3>An Infrastructure Management Tool</h3>
        <p>
            Monad is unlike other IaC tools in that it...
        </p>
        <ul>
            <li>Has no state file</li>
            <li>Has no yaml configuration</li>
            <li>Has finite goals</li>
        </ul>
        <p>
            Monad implements a strict set of infrastructure patterns using a reliable set of naming
            conventions. Traditional IaC concerns requiring stateful caching, locking, or markup language 
            interpretation are eased to the point of elimination due to it being a product with finite
            intentions.
        </p>
    </section>
    <section>
        <h3>Convention Driven</h3>
        <p>All configurations have conventional defaults, for example:</p>
        <ul>
            <li>Service name defaults to directory name</li>
            <li>Git info (owner/repo/branch/SHA) auto-detected from repository</li>
            <li>AWS settings derived from your SDK configuration</li>
            <li>Conservative resource limits (128MB memory, 3s timeout)</li>
        </ul>
    </section>
    <section>
        <h3>Fully Configurable</h3>
        <p>
            There are no required configuration inputs to Monad beyond the basis of Git provenance data
            and AWS credentials. 
        </p>
        <p>
            All optionally provided configurations are either command line flags 
            or AWS API json bodies.
        </p>
    </section>

    <section id="installation">
        <h2>Install</h2>
        
        <h3>GitHub Releases</h3>
        <p>Linux, Mac and Windows binaries are available from <a href="https://github.com/bkeane/monad/releases" target="_blank">GitHub releases</a>.</p>
        
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
        <p>
            Monad aims to be self documenting from the command line. This document intends to provide some
            further support in understanding the pattern of using Monad.
        </p>
    </section>
</template>