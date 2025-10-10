<script setup>
import { RouterLink, RouterView } from 'vue-router'
import { ref, onMounted } from 'vue'
import api from './api/axios'
import HelloWorld from './components/HelloWorld.vue'

const role = ref("") // "", "reader", "writer", "admin"

onMounted(async () => {
  const u = localStorage.getItem("username");
  if (!u) return; // login değilse rol boş kalsın
  try {
    const { data } = await api.get(`/user/${encodeURIComponent(u)}`);
    role.value = data?.data?.role || "";
  } catch (e) {
    // rol alınamazsa sessiz geç; menüde "Blog Oluştur" görünmez
    role.value = "";
    console.warn("role fetch failed:", e?.response?.data || e?.message);
  }
});

function canCreate() {
  return role.value === "writer" || role.value === "admin";
}

function isAdmin() {
  return role.value === "admin";
}
</script>

<template>
  <header>
    <img alt="Vue logo" class="logo" src="@/assets/logo.svg" width="125" height="125" />

    <div class="wrapper">
      <HelloWorld msg="You did it!" />

      <nav>
        <RouterLink to="/">Home</RouterLink>
        <RouterLink to="/me">Hesabım</RouterLink>
        <RouterLink v-if="canCreate()" to="/blog-create">Blog Oluştur</RouterLink>
        <RouterLink to="/users">Kullanıcılar</RouterLink>
        <RouterLink to="/blogs">Bloglar</RouterLink>
        <RouterLink v-if="role === 'admin'" to="/admin/pending">Onay Bekleyenler</RouterLink>
      </nav>
    </div>
  </header>

  <RouterView />
</template>

<style scoped>
header { line-height: 1.5; max-height: 100vh; }
.logo { display: block; margin: 0 auto 2rem; }
nav { width: 100%; font-size: 12px; text-align: center; margin-top: 2rem; }
nav a.router-link-exact-active { color: var(--color-text); }
nav a.router-link-exact-active:hover { background-color: transparent; }
nav a { display: inline-block; padding: 0 1rem; border-left: 1px solid var(--color-border); }
nav a:first-of-type { border: 0; }
@media (min-width: 1024px) {
  header { display: flex; place-items: center; padding-right: calc(var(--section-gap) / 2); }
  .logo { margin: 0 2rem 0 0; }
  header .wrapper { display: flex; place-items: flex-start; flex-wrap: wrap; }
  nav { text-align: left; margin-left: -1rem; font-size: 1rem; padding: 1rem 0; margin-top: 1rem; }
}
</style>