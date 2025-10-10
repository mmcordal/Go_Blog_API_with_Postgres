<script setup>
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import api from "../api/axios";

const router = useRouter();
const user = ref(null);

onMounted(async () => {
  const token = localStorage.getItem("token");
  const username = localStorage.getItem("username");
  if (!token || !username) {
    router.push("/login");
    return;
  }

  try {
    const res = await api.get(`/user/${encodeURIComponent(username)}`);
    // Backend: { data: UserVM }
    user.value = res.data?.data || null;
  } catch (e) {
    // Eğer token expired ise interceptor seni /login'e yönlendirecek
    console.error("Profil alınamadı:", e);
  }
});

const logout = () => {
  localStorage.clear();
  router.push("/login");
};
</script>

<template>
  <div v-if="user">
    <h2>Hoş geldin, {{ user.username }}</h2>
    <p>Email: {{ user.email }}</p>
    <p>Rol: {{ user.role }}</p>
    <button @click="logout">Çıkış Yap</button>
  </div>

  <div v-else>
    <p>Yükleniyor...</p>
  </div>
</template>
