import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import "./assets/main.css";

const app = createApp(App);

// Beyaz ekran yerine hata yakalayıcı
app.config.errorHandler = (err, instance, info) => {
    console.error("Vue error:", err, info);
};

app.use(router).mount("#app");
