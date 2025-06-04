import { createApp } from 'vue'
import App from './App.vue'
import highlight from 'highlight.js/lib/core';
import bash from 'highlight.js/lib/languages/bash';
import dockerfile from 'highlight.js/lib/languages/dockerfile';

highlight.registerLanguage('bash', bash);
highlight.registerLanguage('dockerfile', dockerfile);

createApp(App).mount('#app')
