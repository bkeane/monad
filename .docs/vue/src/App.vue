<template>
  <div class="content">
    <nav>
      <ul>
        <li>
          <h1 class="logo-container">
            <MonadAnimation class="logo-scaled" />
            <span><a href="#/home" class="contrast">Monad</a></span>
          </h1>
        </li>
      </ul>
      <ul>
        <li><a class="contrast" href="#/design">Design</a></li>
        <li><a class="contrast" href="#/usage">Usage</a></li>
        <li><a class="contrast" href="#/dive">Dive</a></li>
        <li><a class="contrast" href="https://github.com/bkeane/monad">Source</a></li>
      </ul>
    </nav>
    <component :is="currentView" />
  </div>
</template>

<script setup>
import '@picocss/pico/css/pico.min.css';
import Home from './components/Home.vue';
import Design from './components/Design.vue';
import Usage from './components/Usage.vue';
import Dive from './components/Dive.vue';
import MonadAnimation from './components/Common/AtomAnimation.vue';
import { ref, computed } from 'vue';

const routes = {
  '/home': Home,
  '/design': Design,
  '/usage': Usage,
  '/dive': Dive
};

const currentPath = ref(window.location.hash)

window.addEventListener('hashchange', () => {
  currentPath.value = window.location.hash
})

const currentView = computed(() => {
  return routes[currentPath.value.slice(1) || '/home'] || NotFound
})
</script>

<style scoped>
.contrast {
    text-decoration: none;
}

.content {
    margin: 0 15%;
}

.logo-container {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin: 0;
}

.logo-scaled {
    transform: scale(0.5);
    transform-origin: left center;
    width: 30%;
    height: 30%;
    display: inline-block;
}
</style>