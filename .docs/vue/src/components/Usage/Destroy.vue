<template>
    <div class="grid">
        <div>
            <h4>Destroy</h4>
            <Code :code="`monad --service <name> destroy`" />
        </div>
        <div>
        </div>
    </div>
</template>

<script>
import Code from '../Common/Code.vue';

export default {
    components: {
        Code
    }
}
</script>
