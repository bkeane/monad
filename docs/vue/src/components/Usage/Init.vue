<template>
        <div class="grid">
            <div>
                <h2>Init</h2>
                <pre><code class="bash">
git init
git add origin https://github.com/{{ git.owner }}/{{ git.repo }}
git checkout -b {{ git.branch }}
monad init {{ `<language>` }}

</code></pre>
            </div>
            <div style="margin-top: 25%; margin-left: 10%;">
                <git-graph />
            </div>
        </div>
</template>

<script>
import GitGraph from '../../../assets/diagrams/git-graph-1.svg?component';

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
        GitGraph
    }
}
</script>
