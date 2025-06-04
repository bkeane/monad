<template>
    <div>
        <h3>IAM</h3>
        <p>
            By default monad will deploy your lambda function with minimum role and policy adequate for execution.
        </p>
        <h5>Custom Policy</h5>
        <p>
            <Code :code="policy" :language="`bash`" />
        </p>
        <p class="doc-ref">
            <a href="https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_statement.html">Policy Structure</a><br>
            <a href="https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-rule-conditions.html">Policy Simulator</a><br>
            <a href="https://docs.aws.amazon.com/service-authorization/latest/reference/reference_policies_actions-resources-contextkeys.html">Actions, Resources & Conditions</a>
        </p>
        <h5>Custom Role Policy</h5>
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
}

.doc-ref a {
    font-style: none;
    color: inherit;
}
</style>