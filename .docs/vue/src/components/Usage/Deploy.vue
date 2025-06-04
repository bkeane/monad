<template>
    <div>
        <h2>Deploy</h2>
        <div v-for="diagram in diagrams" :key="diagram.title">
            <Header :columns="[diagram.title, ``]" />
            <div class="grid">
                <Code :code="diagram.command" :title="diagram.title" />
                <Diagram :png="diagram.png" />
            </div>
        </div>
    </div>
</template>

<script>
import Header from '../Common/Header.vue';
import Code from '../Common/Code.vue';
import Diagram from '../Common/Diagram.vue';
import deployBasic from '../../../assets/diagrams/deployment-basic-1.png';
import deployWebserver from '../../../assets/diagrams/deployment-webserver-1.png';
import deployWorker from '../../../assets/diagrams/deployment-worker-1.png';
import deployCombo from '../../../assets/diagrams/deployment-combo-1.png';

export default {
    components: {
        Header,
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
                    png: deployBasic
                },
                {
                    title: 'Webserver',
                    command: 'monad --service <name> deploy --api $api_gateway_name',
                    gateway: true,
                    rule: false,
                    png: deployWebserver
                },
                {
                    title: 'Worker',
                    command: 'monad --service <name> deploy --rule file://rule.json',
                    gateway: false,
                    rule: true,
                    png: deployWorker
                },
                {
                    title: 'Webserver + Worker',
                    command: 'monad --service <name> deploy --rule file://rule.json --api $api_gateway_name',
                    gateway: true,
                    rule: true,
                    png: deployCombo
                }
            ]
        }
    }
}
</script>
