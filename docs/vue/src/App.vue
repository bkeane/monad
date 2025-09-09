<template>
  <article>
    <h1>Monad</h1>
    <subtitle>GitOps in a Binary</subtitle>
    <component :is="currentView" />
  </article>

  <Nav github="https://github.com/bkeane/monad">
    <ul>
      <li><a href="#/introduction">Introduction</a><ul>
        <li><a href="#/introduction/installation">Install</a></li>
        <li><a href="#/introduction/usage">Usage</a><ul>
        </ul></li>
      </ul></li>

      <li><a href="#/artifact">Artifact</a><ul>
        <li><a href="#/artifact/init">Init</a></li>
        <li><a href="#/artifact/publish">Publish</a></li>
        <li><a href="#/artifact/deploy">Deploy</a></li>
      </ul></li>

      <li><a href="#/permissions">RBAC</a><ul>
        <li><a href="#/permissions/role">Role</a></li>
        <li><a href="#/permissions/policy">Policy</a></li>
        <li><a href="#/permissions/boundary">Boundary</a></li>
      </ul></li>

      <li><a href="#/env">Env</a><ul>
        <li><a href="#/env/dotenv">Dotenv</a></li>
        <li><a href="#/env/entrypoint">Entrypoint</a></li>
      </ul></li>

      <li><a href="#/http">Http</a><ul>
        <li><a href="#/http/api">Api</a></li>
        <li><a href="#/http/route">Route</a></li>
        <li><a href="#/http/auth">Auth</a></li>
      </ul></li>

      <li><a href="#/event">Event</a><ul>
        <li><a href="#/event/bus">Bus</a></li>
        <li><a href="#/event/rule">Rule</a></li>
      </ul></li>

      <li><a href="#/render">Render</a><ul>
        <li><a href="#/render/list">List</a></li>
        <li><a href="#/render/string">String</a></li>
        <li><a href="#/render/file">File</a></li>
      </ul></li>

    </ul>
  </Nav>
</template>

<script setup>
import { ref, computed, nextTick } from 'vue'

// Import page components
import Introduction from './components/Introduction.vue';
import Artifact from './components/Artifact.vue';
import Http from './components/Http.vue';
import Event from './components/Event.vue';
import Rbac from './components/Rbac.vue';
import Env from './components/Env.vue';
import Render from './components/Render.vue';

// Simple client-side routing
const routes = {
  '/introduction': Introduction,
  '/artifact': Artifact,
  '/http': Http,
  '/event': Event,
  '/permissions': Rbac,
  '/env': Env,
  '/render': Render,
}

const currentPath = ref(window.location.hash)

// Function to scroll to element if hash contains an ID
const scrollToId = (hash) => {
  nextTick(() => {http://localhost:5173/monad/#/event/bus
    if (hash.includes('/')) {
      const parts = hash.split('/')
      const id = parts[parts.length - 1]
      if (id && id !== parts[1]) {
        const element = document.getElementById(id)
        if (element) {
          element.scrollIntoView({ behavior: 'smooth' })
        }
      }
    }
  })
}

window.addEventListener('hashchange', () => {
  currentPath.value = window.location.hash
  scrollToId(window.location.hash)
})

// Handle initial load
if (window.location.hash) {
  scrollToId(window.location.hash)
}

const currentView = computed(() => {
  const hash = currentPath.value.slice(1) || '/introduction'
  // Extract the main route (first part after #/)
  const mainRoute = '/' + hash.split('/')[1]
  return routes[mainRoute] || Introduction
})
</script>