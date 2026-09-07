import { createApp } from "vue";
import { createPinia } from "pinia";
import App from "./App.vue";
import { router } from "./router";
import "./style.css";

export const SERVER_URL = `${window.location.protocol}//${window.location.hostname}:${window.location.port}`;

createApp(App).use(createPinia()).use(router).mount("#app");
