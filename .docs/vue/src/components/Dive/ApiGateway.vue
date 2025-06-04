<template>
    <div>
        <h3>API Gateway V2</h3>
        <p>
            To reverse proxy https traffic to your lambda function you can provide it an HTTP API Gateway V2 on which to mount.
        </p>
        <h5>Mount</h5>
        <p>
            <Code :code="mount" :language="`bash`" />
        </p>
        <p class="doc-ref">
            <a href="https://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-set-up-simple-proxy.html">Greedy Proxies</a><br>
            <a href="https://learn.microsoft.com/en-us/aspnet/core/host-and-deploy/proxy-load-balancer?view=aspnetcore-9.0">X-Forwarded-Prefix</a>
        </p>
        <h5>Authorizors</h5>
        <p>
            <Code :code="authorizers" :language="`bash`" />
        </p>
        <p class="doc-ref">
            <a href="https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-access-control-iam.html">IAM Authentication</a><br>
            <a href="https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-jwt-authorizer.html">JWT Authentication</a><br>
            <a href="https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-lambda-authorizer.html">Custom Authentication</a>
        </p>
        <h5>Invoke</h5>
        <p>
            <Code :code="curl" :language="`bash`" />
        </p>
        <p class="doc-ref">
            <a href="https://github.com/okigan/awscurl">AWSCurl</a>
        </p>
    </div>
</template>

<script>
import Code from '../Common/Code.vue';

export default {
    name: 'ApiGateway',
    components: {
        Code
    },
    data() {
        return {
            mount: `monad deploy --api $api_id_or_name`,
            authorizers: `monad deploy --api name --auth $authorizer_id_or_name`,
            curl: "curl https://$api_gateway_url/${repo}/${branch}/${service}/",
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