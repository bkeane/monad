<template>
    <div class="grid">
        <div>
            <h2>Init</h2>
            <Code :code="code" />
        </div>
        <div style="margin-top: 25%; margin-left: 10%;">
            <GitGraph />
        </div>
    </div>
</template>

<script>
import GitGraph from '../../../assets/diagrams/git-graph-1.svg?component';
import Code from '../Common/Code.vue';
import endent from 'endent';

export default {
    props: {
        git: {
            type: {
                owner: String,
                repo: String,
                service: String,
                branch: String,
                sha: String,
            },
            required: true
        },
    },
    components: {
        Code,
        GitGraph
    },
    data() {
        return {
            code: endent`
git init
git add origin https://github.com/${ this.git.owner }/${ this.git.repo }
git checkout -b ${ this.git.branch }
monad init <language>
            `
        }
    }
}
</script>
