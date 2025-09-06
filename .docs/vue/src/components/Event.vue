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
{{`monad deploy --bus $bus`}}
        </CodeBlock>
    </section>
    
    <section id="rule">
        <h2>Rule</h2>
        <p>The rule flag can be used to customize eventing rules:</p>
        <CodeBlock language="bash">
{{`monad deploy --bus $bus --rule ./rule.json.tmpl`}}
        </CodeBlock>
    </section>
    
    <section id="convention">
        <h2>Convention</h2>
        <h3>Rules</h3>
        <p>By default the lambda will be invoked by events matching <code>{ Source: /$repo/$branch/* }</code>.</p>
    </section>
    
    <section id="configuration">
        <h2>Configuration</h2>
        
        <h3>Event Rules</h3>
        <p>At init time you can generate the default rule template for modification:</p>
        <CodeBlock language="bash">
{{`monad init go --rule # output: rule.json.tmpl
monad deploy --bus $bus --rule ./rule.json.tmpl`}}
        </CodeBlock>
        <p>...Or you can create a valid rule json by hand.</p>
    </section>
</template>