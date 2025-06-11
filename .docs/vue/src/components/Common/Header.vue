<template>
    <div class="grid">
        <div v-for="column in columns" :key="column">
            <h6>{{ column }}</h6>
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

