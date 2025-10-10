// src/router/index.js
import { createRouter, createWebHistory } from "vue-router";
import LoginView from "../views/LoginView.vue";
import RegisterView from "../views/RegisterView.vue";
import HomeView from "../views/HomeView.vue";
import UsersView from "../views/UsersView.vue";
import MyAccountView from "../views/MyAccountView.vue";
import BlogCreateView from "../views/BlogCreateView.vue";
import BlogsSearchView from "../views/BlogsSearchView.vue";
import BlogsAllView from "../views/BlogsAllView.vue";
import UserProfileView from "../views/UserProfileView.vue"; // 👈 YENİ
import AdminPendingBlogsView from "../views/AdminPendingBlogsView.vue";


const router = createRouter({
    history: createWebHistory(),
    routes: [
        { path: "/", redirect: "/home" },
        { path: "/login", component: LoginView },
        { path: "/register", component: RegisterView },
        { path: "/home", component: HomeView },
        { path: "/users", component: UsersView },
        { path: "/me", component: MyAccountView },

        // Kullanıcı profili (başkalarının profiline bakma)
        { path: "/u/:username", component: UserProfileView }, // 👈 YENİ
        { path: "/admin/pending", component: AdminPendingBlogsView },


        // Bloglar
        { path: "/blogs", component: BlogsSearchView }, // arama
        { path: "/blogs/all", component: BlogsAllView }, // tüm bloglar
        { path: "/blog-create", component: BlogCreateView },
    ],
});

router.beforeEach((to, from, next) => {
    const token = localStorage.getItem("token");
    const open = ["/login", "/register"];
    if (!open.includes(to.path) && !token) {
        next("/login");
    } else {
        next();
    }
});

export default router;
