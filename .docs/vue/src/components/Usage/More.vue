<template> 
    <div class="grid">
            <div>
                <h2>And more...</h2>
                <Code :code="code" />
            </div>
    <div></div>
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
monad -h
            `
        }
    }
}
</script>