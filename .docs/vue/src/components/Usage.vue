<template>
    <div>
        <h3>Usage</h3>
        <Install :git="git" />
        <Init :git="git" />
        <Publish />
        <Deploy :git="git" :resource="resource" />
        <Destroy />
        <More />
    </div>
</template>

<script>
import Install from './Usage/Install.vue';
import Init from './Usage/Init.vue';
import Publish from './Usage/Publish.vue';
import Deploy from './Usage/Deploy.vue';
import Destroy from './Usage/Destroy.vue';
import More from './Usage/More.vue';

export default {
    components: {
        Install,
        Init,
        Publish,
        Deploy,
        Destroy,
        More
    },
    data() {
        return {
            git: {
                owner: 'owner',
                repo: 'repo',
                branch: 'example',
                sha: 'abd50b21',
                service: 'service',
            },
            resource: {
                name: '',
                path: '',
                image: '',
                logGroup: '',
                route: '',
                tags: {},
            }
        }
    },
    created() {
        this.initializeData();
    },
    methods: {
        initializeData() {
            const { branch, sha, repo, service, owner } = this.git;
            this.resource = {
                name: `${repo}-${branch}-${service}`,
                path: `/${repo}/${branch}/${service}`,
                image: `${owner}/${repo}/${service}:${branch}`,
                logGroup: `/aws/lambda/${repo}/${branch}/${service}`,
                route: `/${repo}/${branch}/${service}/{proxy+}`,
                tags: {
                    "Monad": true,
                    "Branch": branch,
                    "Sha": sha,
                    "Service": service,
                    "Owner": owner,
                    "Repo": repo,
                }
            };
        },
    }
}
</script>