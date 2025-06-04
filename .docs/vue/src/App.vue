<template>
  <div class="content">
    <nav>
      <ul>
        <li><h1>Monad</h1></li>
      </ul>
      <ul>
        <li><a class="contrast" href="#/design">Design</a></li>
        <li><a class="contrast" href="#/usage">Usage</a></li>
        <li><a class="contrast" href="#/dive">Dive</a></li>
      </ul>
    </nav>
    <component :is="currentView" />
  </div>
</template>

<script setup>
import '@picocss/pico/css/pico.min.css';
import Design from './components/Design.vue';
import Usage from './components/Usage.vue';
import Dive from './components/Dive.vue';
import { ref, computed } from 'vue';

const routes = {
  '/design': Design,
  '/usage': Usage,
  '/dive': Dive
};

const currentPath = ref(window.location.hash)

window.addEventListener('hashchange', () => {
  currentPath.value = window.location.hash
})

const currentView = computed(() => {
  return routes[currentPath.value.slice(1) || '/design'] || NotFound
})
</script>

<style scoped>
.content {
    margin: 0 15%;
}
</style>