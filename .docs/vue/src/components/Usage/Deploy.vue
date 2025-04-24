<template>
    <div>
        <h2>Deploy</h2>
        <div v-for="diagram in diagrams" :key="diagram.title">
            <div class="grid">
                <Code :code="diagram.command" :title="diagram.title" />
                <Diagram :content="diagram.svg" />
            </div>
        </div>
    </div>
</template>

<script>
import Code from '../Common/Code.vue';
import Diagram from '../Common/Diagram.vue';
import basicSvg from '../../../assets/diagrams/deployment-basic-1.svg?raw';
import webserverSvg from '../../../assets/diagrams/deployment-webserver-1.svg?raw';
import workerSvg from '../../../assets/diagrams/deployment-worker-1.svg?raw';
import comboSvg from '../../../assets/diagrams/deployment-combo-1.svg?raw';

export default {
    components: {
        Code,
        Diagram
    },
    props: {
        git: {
            type: {
                owner: String,
                repo: String,
                service: String,
                branch: String,
            },
            required: true
        },
        resource: {
            type: {
                name: String,
                path: String,
                image: String,
                logGroup: String,
                tags: Object,
            },
            required: true
        }
    },
    data() {
        return {
            diagrams: [
                {
                    title: 'Function',
                    command: 'monad --service <name> deploy',
                    gateway: false,
                    rule: false,
                    svg: basicSvg
                },
                {
                    title: 'Webserver',
                    command: 'monad --service <name> deploy --api $api_gateway_name',
                    gateway: true,
                    rule: false,
                    svg: webserverSvg
                },
                {
                    title: 'Worker',
                    command: 'monad --service <name> deploy --rule file://rule.json',
                    gateway: false,
                    rule: true,
                    svg: workerSvg
                },
                {
                    title: 'Webserver + Worker',
                    command: 'monad --service <name> deploy --rule file://rule.json --api $api_gateway_name',
                    gateway: true,
                    rule: true,
                    svg: comboSvg
                }
            ]
        }
    }
}
</script>
