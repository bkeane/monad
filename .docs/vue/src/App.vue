<template>
  <div>
    <nav>
      <ul>
        <li><h1>Monad</h1></li>
      </ul>
      <ul>
        <li><a class="contrast" href="#/usage">Usage</a></li>
        <li><a class="contrast" href="#/design">Design</a></li>
        <!-- <li> | </li>
        <li><a class="contrast" href="https://github.com/bkeane/monad" target="_blank">
          <font-awesome-icon :icon="['fab', 'github']" />
        </a></li> -->
      </ul>
    </nav>
    <component :is="currentView" />
  </div>
</template>

<script setup>
import '@picocss/pico/css/pico.min.css';
import Design from './components/Design.vue';
import Usage from './components/Usage.vue';
import { ref, computed, defineComponent } from 'vue'

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

// defineComponent({
//   components: {
//     FontAwesomeIcon
//   }
// })
</script>

<style scoped>
nav {
    margin: 0 15%;
}
</style>