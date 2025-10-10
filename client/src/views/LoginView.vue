<script setup>
import { ref } from "vue";
import api from "../api/axios";
import { useRouter } from "vue-router";

const router = useRouter();
const identifier = ref("");
const password = ref("");

// modal state
const showModal = ref(false);
const modalMessage = ref("");

const login = async () => {
  try {
    const res = await api.post("/login", {
      identifier: identifier.value,
      password: password.value,
    });

    const payload = res.data?.data || {};
    localStorage.setItem("token", payload.token);
    localStorage.setItem("username", payload.username);
    localStorage.setItem("email", payload.email);
    localStorage.setItem("id", String(payload.id));

    router.push("/home");
  } catch (err) {
    modalMessage.value =
        err?.response?.data?.error ||
        "Login başarısız! Kullanıcı adı/email veya şifre hatalı.";
    showModal.value = true; // modal aç
  }
};

const closeModal = () => {
  showModal.value = false; // sadece modal kapanır, login form ekranda kalır
};
</script>

<template>
  <div>
    <h2>Giriş Yap</h2>
    <input v-model="identifier" placeholder="Kullanıcı adı veya Email" />
    <input v-model="password" type="password" placeholder="Şifre" />
    <button @click="login">Giriş Yap</button>

    <router-link to="/register">Kayıt ol</router-link>

    <!-- Modal -->
    <div v-if="showModal" class="modal">
      <div class="modal-content">
        <p>{{ modalMessage }}</p>
        <button @click="closeModal">Tamam</button>
      </div>
    </div>
  </div>
</template>

<style>
.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  justify-content: center;
  align-items: center;
  background: rgba(0,0,0,0.5);
}
.modal-content {
  background: #46a153;
  padding: 20px;
  border-radius: 8px;
  max-width: 400px;
  text-align: center;
}
</style>
