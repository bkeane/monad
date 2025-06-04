<template>
    <div>
        <h2>Dive</h2>
        <OCI />
        <Environment />
        <IAM />
        <EventBridge />
        <ApiGateway />
    </div>
</template>

<script>
import IAM from './Dive/IAM.vue';
import EventBridge from './Dive/EventBridge.vue';
import ApiGateway from './Dive/ApiGateway.vue';
import Environment from './Dive/Environment.vue';
import OCI from './Dive/OCI.vue';

export default {
    name: 'Dive',
    components: {
        IAM,
        EventBridge,
        ApiGateway,
        Environment,
        OCI
    }
}
</script>