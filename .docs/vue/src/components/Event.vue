<script setup>
import { CodeBlock, Figure } from 'vue-tufte';
import worker from '../../assets/diagrams/deployment-worker-1.png';
</script>

<template>
    <section id="requirements">
        <h2>Requirements</h2>
        <p>Event-driven functions are deployed as Lambda functions with EventBridge rules.</p>
    </section>

    <Figure :src="worker"></Figure>
    
    <section id="bus">
        <h2>Bus</h2>
        <p>The bus flag wires the lambda up to a named bus:</p>
        <CodeBlock language="bash">
{{`monad deploy --bus production`}}
        </CodeBlock>
    </section>
    
    <section id="subscribe">
        <h2>Subscribe</h2>
        <p>The rule flag can be used to customize event filter rules:</p>
        <CodeBlock language="bash">
{{`monad deploy --bus default --rule rule.json.tmpl`}}
        </CodeBlock>
    </section>
    
    <section id="publish">
        <h2>Publish</h2>
        <p>Functions can publish events to EventBridge buses using the AWS SDK.</p>
    </section>
    
    <section id="convention">
        <h2>Convention</h2>
        <h3>Subscribe</h3>
        <p>By default the lambda will be invoked by events with...</p>
        <ul>
          <li>Source: `/{repo}/{branch}/*`</li>
          <li>!Source: `/{repo}/{branch}/{service}`</li>
        </ul>
        <p>ie: it will be invoked by events within the repository/branch namespace, but not by iteself.</p>
        <h3>Publish</h3>
        <p>There is no permission allowing eventbridge emit by default.</p>
    </section>
    
    <section id="configuration">
        <h2>Configuration</h2>
        
        <h3>Event Rules</h3>
        <p>Define EventBridge rules using cron strings or JSON templates:</p>
        <CodeBlock language="bash">
{{`# Schedule-based (cron/rate)
monad deploy --bus production --rule "rate(5 minutes)"

# Event pattern (JSON file)
monad deploy --bus production --rule ./event-pattern.json
`}}
        </CodeBlock>
    </section>
</template>