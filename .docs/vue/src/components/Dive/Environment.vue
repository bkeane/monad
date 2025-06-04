<template>
    <div>
        <h3>Environment</h3>
        <h5>Non-Sensitive</h5>
        <p>
            A dotenv file can be provided to set non-sensitive environment variables.
        </p>
        <p>
            <Code :code="dotenv" :language="`bash`" />
        </p>
        <h5>Sensitive</h5>
        <p>
            Sensitive environment variables should be provided by alternate means at runtime.
        </p>
        <p>
            <Code :code="df" :language="`dockerfile`" />
        </p>
        <p class="doc-ref">
            <a href="https://git.sr.ht/~raymond/mapenv">MapEnv</a><br>
            <a href="https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html">Parameter Store</a><br>
            <a href="https://docs.aws.amazon.com/secretsmanager/latest/userguide/intro.html">Secrets Manager</a>
        </p>
    </div>
</template>

<script>
import endent from 'endent';
import Code from '../Common/Code.vue';

export default {
    name: 'Environment',
    components: {
        Code
    },
    data() {
        return {
            dotenv: `monad deploy --env file://.env`,
            df: endent`
COPY --from=raylas/mapenv:latest /bin/mapenv /bin/mapenv
ENTRYPOINT ["/bin/mapenv", "ssm", "-p", "/path/to/env", "--"]
CMD ["./webserver"]
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