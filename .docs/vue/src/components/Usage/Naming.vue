<template>
    <details>
        <summary role="button" class="outline secondary">
            result
        </summary>
        <div class="grid">
            <div>
                <h5>AWS Resources</h5>
                <table>
                    <thead>
                        <tr>
                            <th>resource</th>
                            <th>name</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <td>Image</td>
                            <td><code class="bash">{{ resource.image }}</code>
                            </td>
                        </tr>
                        <tr>
                            <td>Lambda</td>
                            <td><code class="bash">{{ resource.name }}</code>
                            </td>
                        </tr>
                        <tr>
                            <td>Role</td>
                            <td><code class="bash">{{ resource.name }}</code>
                            </td>
                        </tr>
                        <tr>
                            <td>Policy</td>
                            <td><code class="bash">{{ resource.name }}</code>
                            </td>
                        </tr>
                        <tr>
                            <td>Log Group</td>
                            <td><code class="bash">{{ resource.logGroup }}</code></td>
                        </tr>
                        <tr v-if="rule">
                            <td>Rule</td>
                            <td><code class="bash">{{ resource.name }}</code>
                            </td>
                        </tr>
                        <tr v-if="gateway">
                            <td>Route</td>
                            <td><code class="bash">{{ resource.route }}</code></td>
                        </tr>
                    </tbody>
                </table>
            </div>
            <div>
                <h5>Resource Tags</h5>
                <pre><code class="json">{{ resource.tags }}</code></pre>
            </div>
        </div>
    </details>
</template>

<script>
export default {
    props: {
        gateway: {
            type: Boolean,
            default: false
        },
        rule: {
            type: Boolean,
            default: false
        },
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
        resource: {
            type: {
                name: String,
                path: String,
                image: String,
                logGroup: String,
                route: String,
                tags: Object,
            },
            required: true
        }
    }
}
</script>