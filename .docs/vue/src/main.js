import { createApp } from 'vue'
import App from './App.vue'
import highlight from 'highlight.js/lib/core';
import bash from 'highlight.js/lib/languages/bash';

highlight.registerLanguage('bash', bash);

createApp(App).mount('#app')
