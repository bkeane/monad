<template>
    <div>
        <h4>Install</h4>
        <h6>Homebrew (<a href="https://github.com/bkeane/monad/releases">releases</a>)</h6>
        <div class="grid">
            <Code :code="brew" />
            <div></div>
        </div>
        <h6>Github Actions (<a href="https://github.com/marketplace/actions/monad-action">releases</a>)</h6>
        <div class="grid">
            <Code :code="action" />
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
            brew: endent`
brew tap bkeane/tools
brew install bkeane/tools/monad
            `,
            action: endent`
- uses: bkeane/monad-action@v1
            `
        }
    }
}
</script>