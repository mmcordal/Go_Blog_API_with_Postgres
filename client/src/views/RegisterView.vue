<script setup>
import { ref } from "vue";
import api from "../api/axios";
import { useRouter } from "vue-router";

const router = useRouter();
const username = ref("");
const email = ref("");
const password = ref("");
const role = ref("reader");
const error = ref("");

// modal state
const showModal = ref(false);
const successMessage = ref("");

const register = async () => {
  try {
    error.value = "";
    const res = await api.post("/register", {
      username: username.value,
      email: email.value,
      password: password.value,
      role: role.value,
    });

    successMessage.value = res.data?.message || "Kayıt başarılı!";
    showModal.value = true; // modal aç
  } catch (err) {
    error.value = err?.response?.data?.error || "Kayıt başarısız!";
  }
};

const closeModal = () => {
  showModal.value = false;
  router.push("/login"); // modal kapatılınca login'e yönlendir
};
</script>

<template>
  <div>
    <h2>Kayıt Ol</h2>
    <input v-model="username" placeholder="Kullanıcı adı" />
    <input v-model="email" placeholder="Email" />
    <input v-model="password" type="password" placeholder="Şifre" />

    <select v-model="role">
      <option value="reader">Reader</option>
      <option value="writer">Writer</option>
      <option value="admin">Admin</option>
    </select>

    <button @click="register">Kayıt Ol</button>

    <p v-if="error" style="color:red">{{ error }}</p>

    <router-link to="/login">Giriş yap</router-link>

    <!-- Modal -->
    <div v-if="showModal" class="modal">
      <div class="modal-content">
        <p>{{ successMessage }}</p>
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
  background: #000000;
  padding: 20px;
  border-radius: 8px;
  max-width: 400px;
  text-align: center;
}
</style>
