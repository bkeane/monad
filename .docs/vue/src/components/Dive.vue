<template>
    <div>
        <h3>Dive</h3>
        <Artifact />
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
import Artifact from './Dive/Artifact.vue';

export default {
    name: 'Dive',
    components: {
        IAM,
        EventBridge,
        ApiGateway,
        Environment,
        Artifact
    }
}
</script>