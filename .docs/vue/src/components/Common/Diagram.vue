<template>
    <div>
        <img :src="png" />
    </div>
</template>

<script>
export default {
    props: {
        png: {
            type: String,
            required: true
        },
    }
}
</script>

<style scoped>
@import 'highlight.js/styles/default.css';
</style>
