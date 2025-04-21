<script setup>
import '@picocss/pico/css/pico.min.css';
import Design from './components/Design.vue';
import Usage from './components/Usage.vue';
import { ref, computed } from 'vue'

const routes = {
  '/design': Design,
  '/usage': Usage
}

const currentPath = ref(window.location.hash)

window.addEventListener('hashchange', () => {
  currentPath.value = window.location.hash
})

const currentView = computed(() => {
  return routes[currentPath.value.slice(1) || '/usage'] || NotFound
})

</script>

<template>
  <div>
    <nav>
      <ul>
        <li><strong>Monad</strong></li>
      </ul>
      <ul>
        <li><a href="#/usage">Usage</a></li>
        <li><a href="#/design">Design</a></li>
      </ul>
    </nav>
    <component :is="currentView" />
  </div>
</template>
