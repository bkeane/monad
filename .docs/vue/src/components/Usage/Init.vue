<template>
    <div>
        <h4>Init</h4>
        <div class="grid">
            <Code :code="code" />
            <Diagram :png="png" />
        </div>
    </div>
</template>

<script>
import Header from '../Common/Header.vue';
import Code from '../Common/Code.vue';
import Diagram from '../Common/Diagram.vue';
import gitGraph from '../../../assets/diagrams/git-graph-1.png';
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
        Header,
        Code,
        Diagram
    },
    data() {
        return {
            png: gitGraph,
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
