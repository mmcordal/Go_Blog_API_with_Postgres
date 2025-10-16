import { createRouter, createWebHistory } from "vue-router";
import api from "../api/axios";

import LoginView from "../views/LoginView.vue";
import RegisterView from "../views/RegisterView.vue";
import HomeView from "../views/HomeView.vue";
import UsersView from "../views/UsersView.vue";
import MyAccountView from "../views/MyAccountView.vue";
import BlogCreateView from "../views/BlogCreateView.vue";
import BlogsSearchView from "../views/BlogsSearchView.vue";
import BlogsAllView from "../views/BlogsAllView.vue";
import UserProfileView from "../views/UserProfileView.vue";

// Admin sayfaları (lazy)
const AdminPendingBlogsView = () => import("../views/AdminPendingBlogsView.vue");
const AdminRoleRequestsView = () => import("../views/AdminRoleRequestsView.vue");

// 🔵 ÖNEMLİ: base'i açıkça "/" yapıyoruz ki /blog-create gibi bir alt-yol eklenmesin
const router = createRouter({
    history: createWebHistory("/"),
    routes: [
        { path: "/", redirect: "/home" },
        { path: "/login", component: LoginView },
        { path: "/register", component: RegisterView },
        { path: "/home", component: HomeView },
        { path: "/users", component: UsersView },
        { path: "/me", component: MyAccountView },
        { path: "/u/:username", component: UserProfileView },
        { path: "/blogs", component: BlogsSearchView },
        { path: "/blogs/all", component: BlogsAllView },
        { path: "/blog-create", component: BlogCreateView },

        // admin
        { path: "/admin/pending", component: AdminPendingBlogsView, meta: { requiresAdmin: true } },
        { path: "/admin/role-requests", component: AdminRoleRequestsView, meta: { requiresAdmin: true } },

        { path: "/:pathMatch(.*)*", redirect: "/home" },
    ],
});

router.beforeEach(async (to, from, next) => {
    const token = localStorage.getItem("token");
    const publicPages = ["/login", "/register", "/home", "/blogs", "/blogs/all"];

    // public sayfalar hariç token iste
    if (!publicPages.includes(to.matched[0]?.path) && !token) {
        if (to.path !== "/login") return next("/login");
    }

    // admin kontrolü
    if (to.matched.some(r => r.meta?.requiresAdmin)) {
        let role = localStorage.getItem("role");
        if (!role && token) {
            try {
                const { data } = await api.get("/me");
                role = data?.data?.role || "";
                if (role) localStorage.setItem("role", role);
            } catch {
                role = "";
            }
        }
        if (role !== "admin") {
            alert("Bu sayfaya sadece admin erişebilir.");
            return next("/home");
        }
    }

    next();
});

export default router;
