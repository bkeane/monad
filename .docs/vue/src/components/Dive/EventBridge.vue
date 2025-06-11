<template>
    <div>
        <h4>EventBridge</h4>
        <p>
            To wire up your lambda function to be invoked by eventbridge events you can provide both a bus and rule.
        </p>
        <h6>Rule</h6>
        <p>
            <Code :code="rule" :language="`bash`" />
        </p>
        <p class="doc-ref">
            <a href="https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-rule-conditions.html">Rule Documentation</a><br>
            <a href="https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-pattern-sandbox.html">Rule Sandbox</a>
        </p>
        <h6>Bus</h6>
        <p>
            <Code :code="bus" :language="`bash`" />
        </p>
        <p class="doc-ref">
            <a href="https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-bus.html">Bus Documentation</a>
        </p>
    </div>
</template>

<script>    
import Code from '../Common/Code.vue';

export default {
    name: 'EventBridge',
    components: {
        Code
    },
    data() {
        return {
            rule: `monad deploy --rule file://rule.json`,
            bus: `monad deploy --bus myBus --rule file://rule.json`
        }
    }
}
</script>

<style scoped>
.doc-ref {
    text-align: right;
    font-size: 0.9rem;
}

.doc-ref a {
    font-style: none;
    color: inherit;
}
</style>