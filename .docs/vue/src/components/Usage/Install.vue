<template>
    <div>
        <h2>Install</h2>
        <h4>Homebrew (or <a href="https://github.com/bkeane/monad/releases">Releases</a>)</h4>
        <div class="grid">
            <Code :code="code" />
            <div></div>
        </div>
    </div>
</template>

<script>
import Code from '../Common/Code.vue';
import endent from 'endent';

export default {
    components: {
        Code
    },
    data() {
        return {
            code: endent`
brew tap bkeane/tools
brew install bkeane/tools/monad
            `
        }
    }
}
</script>