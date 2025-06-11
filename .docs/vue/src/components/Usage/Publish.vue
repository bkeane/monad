<template>
    <div class="grid">
        <div>
            <h4>Publish</h4>
            <Code :code="code" />
        </div>
        <div>
        </div>
    </div>
</template>

<script>
import Code from '../Common/Code.vue';
import endent from 'endent';

export default {
    components: { Code },
    data() {
        return {
            code: endent`
docker build -t $(monad --service <name> ecr tag) --push .`
        }
    }
}
</script>
