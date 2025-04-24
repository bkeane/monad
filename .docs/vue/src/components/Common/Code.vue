<template>
    <div>
        <h4 v-if="title">{{ title }}</h4>
        <pre><code :class="language">
{{ code }}
        </code></pre>
    </div>
</template>

<script>
import highlight from 'highlight.js/lib/core';

export default {
    props: {
        title: {
            type: String,
            required: false,
        },
        language: {
            type: String,
            required: false,
            default: 'bash'
        },
        code: {
            type: String,
            required: true,
        }
    },
    mounted() {
        this.highlightCode();
    },
    watch: {
        code() {
            this.highlightCode();
        }
    },
    methods: {
        highlightCode() {
            const blocks = this.$el.querySelectorAll('pre code');
            blocks.forEach((block) => {
                highlight.highlightElement(block);
            });
        }
    }
}
</script>

<style>
@import 'highlight.js/styles/a11y-dark.css';
</style>