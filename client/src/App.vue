<script setup>
import { RouterLink, RouterView, useRoute } from "vue-router";
import { ref, onMounted, onUnmounted, watch } from "vue";
import api from "./api/axios";

const role = ref("");
const route = useRoute();

async function loadRole() {
  const token = localStorage.getItem("token");
  if (!token) {
    role.value = "";
    return;
  }

  try {
    const { data } = await api.get("/me");
    role.value = data?.data?.role || "";
  } catch (e) {
    // 401 durumunda tekrar login'e zorlamayalım, sadece console.warn
    console.warn("Role yüklenemedi:", e?.response?.data || e?.message);
  }
}

onMounted(() => {
  // login sonrası token yazılmışsa bir nebze bekleyelim
  setTimeout(loadRole, 200);
  window.addEventListener("auth:changed", loadRole);
});

onUnmounted(() => {
  window.removeEventListener("auth:changed", loadRole);
});

watch(() => route.fullPath, () => {
  // her rota değişiminde yeniden role çekmek güvenli
  loadRole();
});

function canCreate() {
  return role.value === "writer" || role.value === "admin";
}
function isAdmin() {
  return role.value === "admin";
}
function logout() {
  localStorage.clear();
  window.dispatchEvent(new Event("auth:changed"));
  window.location.href = "/login";
}
function isLoginPage() {
  return route.path === "/login" || route.path === "/register";
}
</script>

<template>
  <header class="site-header">
    <div class="container header-inner">
      <div class="brand">
        <RouterLink class="nav-link" to="/">
          <img alt="LogNode" class="logo" src="@/assets/logo-lognode.svg" width="140" height="40" />
        </RouterLink>
        <span class="brand-name">LogNode</span>
      </div>

      <nav class="navbar">
        <div v-if="!isLoginPage()">
          <RouterLink vclass="nav-link" to="/">Home</RouterLink>
          <RouterLink class="nav-link" to="/me">Hesabım</RouterLink>
          <RouterLink class="nav-link" v-if="canCreate()" to="/blog-create">Blog Oluştur</RouterLink>
          <RouterLink class="nav-link" to="/users">Kullanıcılar</RouterLink>
          <RouterLink class="nav-link" to="/blogs">Bloglar</RouterLink>
          <RouterLink class="nav-link" v-if="isAdmin()" to="/admin/pending">Onay Bekleyenler</RouterLink>
          <RouterLink class="nav-link" v-if="isAdmin()" to="/admin/role-requests">Rol Talepleri</RouterLink>
          <button class="logout-btn"@click="logout">Çıkış Yap</button>
        </div>
        <div class="social-links">
          <a href="https://github.com/mmcordal/blog-api-go" target="_blank" rel="noopener" class="social-icon">
            <img src="@/assets/github.svg" alt="GitHub" />
          </a>
          <a href="https://linkedin.com/in/melih-cordal" target="_blank" rel="noopener" class="social-icon">
            <img src="@/assets/linkedin.svg" alt="LinkedIn" />
          </a>
        </div>
      </nav>
    </div>
  </header>

  <main class="container page-body">
    <RouterView />
  </main>
</template>

<style scoped>
/* mevcut stiller birebir korundu */
.site-header {
  position: sticky;
  top: 0;
  z-index: 50;
  background: var(--color-background);
  border-bottom: 1px solid var(--color-border);
  backdrop-filter: saturate(120%) blur(6px);
}
.header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 0;
}
.brand { display: flex; align-items: center; gap: 10px; }
.logo { width: 28px; height: 28px; }
.brand-name { font-weight: 700; letter-spacing: .2px; color: var(--color-heading); }
.navbar { display: flex; align-items: center; gap: 14px; justify-content: center; }
.nav-link {
  padding: 8px 10px;
  border-radius: 8px;
  border: 1px solid transparent;
  text-decoration: none;
  color: var(--color-text);
  transition: all .2s ease;
}
.nav-link:hover {
  border-color: var(--color-border-hover);
  background: var(--color-background-soft);
}
.nav-link.router-link-exact-active {
  border-color: hsla(160, 100%, 37%, 0.45);
  box-shadow: 0 0 0 2px hsla(160, 100%, 37%, 0.18) inset;
}
.page-body { padding: 20px 0 32px; }
@media (max-width: 700px) {
  .header-inner { flex-direction: column; gap: 10px; padding: 10px 0; }
  .navbar { flex-wrap: wrap; gap: 8px; }
}
.social-links {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  margin-left: 10px;
  border-left: 1px solid var(--color-border);
  padding-left: 10px;
}
.social-icon img {
  width: 20px;
  height: 20px;
  opacity: 0.8;
  transition: opacity 0.2s, transform 0.2s;
}
.social-icon:hover img {
  opacity: 1;
  transform: scale(1.1);
}
.logout-btn {
  border: 1px solid #19d27c;
  background: transparent;
  color: #19d27c;
  border-radius: 8px;
  padding: 6px 10px;
  cursor: pointer;
  transition: 0.2s ease;
}
.logout-btn:hover {
  background: #19d27c;
  color: black;
}
</style>
