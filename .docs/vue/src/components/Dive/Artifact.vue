<template>
    <div>
        <h4>Artifact</h4>
        <p>
            The deployable artifact of Monad is an OCI / Docker image hosted on AWS ECR.
        </p>
        <h5>Tag</h5>
        <p>
            By default, Monad assumes the tag of your deployable artifact to follow:
        </p>
        <p>
            <Code :code="tag" :language="`bash`" />
        </p>
        <h5>Entrypoint</h5>
        <p>
            All service scaffolds provided by <code>monad init</code> implement a language agnostic pattern
            which assumes the entrypoint of your application is a webserver. This approach is not strictly necessary. 
            Monad only holds this as an opinionated core pattern.
        </p>
        <p>
            <Code :code="entrypoint" :language="`dockerfile`" />
        </p>
        <p class="doc-ref">
            <a href="https://github.com/awslabs/aws-lambda-web-adapter">AWS Lambda Web Adapter</a>
        </p>
    </div>
</template>

<script>
import endent from 'endent';
import Code from '../Common/Code.vue';

export default {
    name: 'Conventions',
    components: {
        Code
    },
    data() {
        return {
            tag: "${id}.dkr.ecr.${region}.amazonaws.com/${owner}/${repo}/${service}:${branch}",
            entrypoint: endent`
            COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.8.4 /lambda-adapter /opt/extensions/lambda-adapter
            ENTRYPOINT ["./webserver"]
            `
        }
    }
}
</script>

<style scoped>
.doc-ref {
    text-align: right;
}

.doc-ref a {
    font-style: none;
    color: inherit;
}
</style>