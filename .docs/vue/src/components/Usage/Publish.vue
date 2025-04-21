<template>
    <div class="grid">
        <div>
            <h2>Publish</h2>
            <pre><code class="bash">
{{`docker build -t $(monad ecr tag <service>) --push`}}
        </code></pre>
        </div>
        <div>
        </div>
    </div>
</template>
