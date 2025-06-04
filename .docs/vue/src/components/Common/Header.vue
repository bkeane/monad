<template>
    <div class="grid">
        <div v-for="column in columns" :key="column">
            <h4>{{ column }}</h4>
        </div>
    </div>
</template>

<script>
export default {
    props: {
        columns: {
            type: Array,
            required: true
        },
    }
}
</script>

