<template>
    <div>
        <h4>IAM</h4>
        <p>
            By default monad will deploy your lambda function with minimum role and policy adequate for execution.
        </p>
        <h6>Custom Policy</h6>
        <p>
            <Code :code="policy" :language="`bash`" />
        </p>
        <p class="doc-ref">
            <a href="https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_statement.html">Policy Structure</a><br>
            <a href="https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-rule-conditions.html">Policy Simulator</a><br>
            <a href="https://docs.aws.amazon.com/service-authorization/latest/reference/reference_policies_actions-resources-contextkeys.html">Actions, Resources & Conditions</a>
        </p>
        <h6>Custom Role Policy</h6>
        <p>
            <Code :code="trust" :language="`bash`" />
        </p>
        <p class="doc-ref">
            <a href="https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles.html">Role Documentation</a>
        </p>
    </div>
</template>

<script>
import Code from '../Common/Code.vue';

export default {
    name: 'Policy',
    components: {
        Code
    },
    data() {
        return {
            policy: `monad deploy --policy file://policy.json`,
            trust: `monad deploy --role file://trust.json`
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