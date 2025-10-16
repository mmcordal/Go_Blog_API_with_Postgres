import axios from "axios";

const base = import.meta.env.VITE_API_BASE_URL || "http://localhost:3000";

const api = axios.create({
    baseURL: `${base}/api/v1`,
});

// Request interceptor
api.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem("token");
        if (token && !config.url.includes("/register") && !config.url.includes("/login")) {
            config.headers = config.headers || {};
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => Promise.reject(error)
);

// Response interceptor (401 → logout + rol yenile tetikle)
api.interceptors.response.use(
    (res) => res,
    (err) => {
        const status = err?.response?.status;
        if (status === 401) {
            localStorage.removeItem("token");
            localStorage.removeItem("username");
            localStorage.removeItem("email");
            localStorage.removeItem("id");

            // 🔔 rol/menu’yu anında güncelle
            try { window.dispatchEvent(new Event("auth:changed")); } catch {}

            if (window.location.pathname !== "/login") {
                window.location.href = "/login";
            }
        }
        return Promise.reject(err);
    }
);

// Log (opsiyonel)
api.interceptors.request.use((config) => {
    console.log("[REQ]", api.defaults.baseURL, config.url);
    return config;
});

export default api;
