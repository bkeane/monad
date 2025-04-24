<template>
    <div>
        <h4 v-if="title">{{ title }}</h4>
        <div v-html="content" />
    </div>
</template>

<script>
export default {
    props: {
        content: {
            type: String,
            required: true
        },
        title: {
            type: String,
            required: false,
        }
    }
}
</script>

<style scoped>
@import 'highlight.js/styles/default.css';

div>div {
    margin-top: 7%;
    margin-bottom: -5%;
    float: center;
}
</style>
