<script setup>
import { ref, computed } from "vue";
import api from "../api/axios";
import { useRouter } from "vue-router";

const router = useRouter();

const username = ref("");
const email = ref("");
const password = ref("");
const role = ref("reader");
const error = ref("");

const showModal = ref(false);
const successMessage = ref("");
const loading = ref(false);

// Alan bazlı hatalar
const errors = ref({
  username: "",
  email: "",
  password: "",
});

// Basit kontroller
function validateUsername(v) {
  const s = v.trim();
  if (!s) return "Kullanıcı adı gerekli";
  if (s.length < 3) return "En az 3 karakter";
  if (s.length > 16) return "En fazla 16 karakter";
  if (!/^[a-zA-Z0-9_]+$/.test(s)) return "Sadece harf, rakam ve alt çizgi";
  return "";
}
function validateEmail(v) {
  const s = v.trim();
  if (!s) return "Email gerekli";
  // hafif e-posta kontrolü
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(s)) return "Geçerli bir email girin";
  return "";
}
function validatePassword(v) {
  if (!v) return "Şifre gerekli";
  if (v.length < 8) return "En az 8 karakter";
  return "";
}

function validateAll() {
  errors.value.username = validateUsername(username.value);
  errors.value.email = validateEmail(email.value);
  errors.value.password = validatePassword(password.value);
  // true -> form geçerli
  return !errors.value.username && !errors.value.email && !errors.value.password;
}

const formInvalid = computed(() => {
  return (
      !username.value.trim() ||
      !email.value.trim() ||
      !password.value ||
      errors.value.username ||
      errors.value.email ||
      errors.value.password
  );
});

async function register() {
  if (loading.value) return;

  // Alanları önce normalize et
  username.value = username.value.trim();
  email.value = email.value.trim();
  role.value = (role.value || "reader").toLowerCase().trim();

  if (!validateAll()) return;

  try {
    loading.value = true;
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
  } finally {
    loading.value = false;
  }
}

function closeModal() {
  showModal.value = false;
  router.push("/login");
}

// Alanlarda yazarken anlık validasyon
function onInputUsername() {
  errors.value.username = validateUsername(username.value);
}
function onInputEmail() {
  errors.value.email = validateEmail(email.value);
}
function onInputPassword() {
  errors.value.password = validatePassword(password.value);
}
</script>

<template>
  <section class="auth-wrap">
    <div class="bg-glow" aria-hidden="true"></div>

    <div class="auth-card">
      <header class="auth-header">
        <img src="@/assets/logo-lognode.svg" alt="LogNode" class="brand" />
        <h1>Kayıt Ol</h1>
        <p class="subtitle">
          LogNode’a hoş geldin. Bir dakikada hesabını oluştur.
        </p>
      </header>

      <div class="form">
        <label for="username">Kullanıcı adı</label>
        <input
            id="username"
            v-model="username"
            placeholder="kullanici_adi"
            autocomplete="username"
            @input="onInputUsername"
            @keyup.enter="register"
        />
        <div v-if="errors.username" class="field-error">{{ errors.username }}</div>

        <label for="email">Email</label>
        <input
            id="email"
            v-model="email"
            type="email"
            placeholder="ornek@mail.com"
            autocomplete="email"
            @input="onInputEmail"
            @keyup.enter="register"
        />
        <div v-if="errors.email" class="field-error">{{ errors.email }}</div>

        <label for="password">Şifre</label>
        <input
            id="password"
            v-model="password"
            type="password"
            placeholder="••••••••"
            autocomplete="new-password"
            @input="onInputPassword"
            @keyup.enter="register"
        />
        <div v-if="errors.password" class="field-error">{{ errors.password }}</div>

        <label for="role">Rol</label>
        <select id="role" v-model="role">
          <option value="reader">Reader</option>
          <option value="writer">Writer</option>
          <!--<option value="admin">Admin</option>-->
        </select>

        <!--<p class="hint">
          <strong>Not:</strong> “Admin” seçilirse hesabın <em>reader</em> olarak
          oluşturulur ve admin olma isteğin yöneticilere iletilir.
        </p>-->

        <router-link to="/login">
        <button class="btn" :disabled="loading || formInvalid" @click="register">
          {{ loading ? "Gönderiliyor..." : "Kayıt Ol" }}
        </button>
        </router-link>

        <p v-if="error" class="error">{{ error }}</p>

        <div class="foot">
          Zaten hesabın var mı?
          <router-link to="/login" class="link">Giriş yap</router-link>
        </div>
      </div>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="modal" @click.self="closeModal">
      <div class="modal-content">
        <h3>🎉 Başarılı</h3>
        <p>{{ successMessage }}</p>
        <button class="btn" @click="closeModal">Tamam</button>
      </div>
    </div>
  </section>
</template>

<style scoped>
/* Sayfa zemin / merkezleme */
.auth-wrap {
  position: relative;
  min-height: calc(100vh - 120px);
  display: grid;
  place-items: center;
  padding: 24px 16px;
  isolation: isolate;
}

/* Yumuşak mint parıltı */
.bg-glow {
  position: absolute;
  inset: 0;
  background:
      radial-gradient(900px 500px at 20% -10%, rgba(25,210,124,0.15), transparent 60%),
      radial-gradient(700px 400px at 85% 10%, rgba(25,210,124,0.10), transparent 60%);
  filter: blur(1px);
  z-index: -1;
}

/* Kart */
.auth-card {
  width: 100%;
  max-width: 520px;
  background: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: 16px;
  box-shadow: 0 10px 50px rgba(0,0,0,0.35);
  padding: 22px 22px 18px;
}

/* Üst bölüm */
.auth-header {
  display: grid;
  gap: 6px;
  margin-bottom: 12px;
  text-align: center;
}
.brand {
  width: 44px;
  height: 44px;
  margin: 0 auto 4px;
  display: block;
}
h1 {
  font-size: 22px;
  font-weight: 700;
  color: var(--color-heading);
}
.subtitle {
  font-size: 13px;
  color: var(--color-text);
  opacity: .8;
}

/* Form */
.form {
  display: grid;
  gap: 10px;
  margin-top: 8px;
}

label {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-heading);
  margin-top: 6px;
}

input,
select,
textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--color-border);
  background: var(--color-background-soft);
  color: var(--color-text);
  border-radius: 10px;
  outline: none;
  transition: border-color .18s ease, box-shadow .18s ease, background .18s ease;
}

input:focus,
select:focus,
textarea:focus {
  border-color: var(--color-border-hover);
  box-shadow: 0 0 0 2px rgba(25,210,124,0.14) inset;
  background: var(--color-background);
}

/* Alan hataları */
.field-error {
  color: #ff5e5e;
  font-size: 12px;
  margin-top: -4px;
}

/* İpucu */
/*.hint {
  font-size: 12px;
  line-height: 1.4;
  color: var(--color-text);
  opacity: .82;
  margin-top: 2px;
}*/

/* Buton */
.btn {
  margin-top: 6px;
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #19d27c;
  border-radius: 10px;
  background: linear-gradient(135deg, #19d27c, #0bbf68);
  color: #0d1b14;
  font-weight: 700;
  cursor: pointer;
  transition: transform .08s ease, filter .18s ease, box-shadow .18s ease;
}
.btn:hover {
  filter: brightness(1.04);
  box-shadow: 0 8px 24px rgba(25,210,124,0.18);
}
.btn:disabled {
  opacity: .7;
  cursor: default;
  box-shadow: none;
}

/* Hata & alt kısım */
.error {
  margin-top: 8px;
  color: #ff5e5e;
  font-size: 13px;
}
.foot {
  margin-top: 10px;
  text-align: center;
  font-size: 13px;
  opacity: .9;
}
.link {
  color: #19d27c;
  text-decoration: none;
  border-bottom: 1px dashed rgba(25,210,124,.35);
  padding-bottom: 2px;
}
.link:hover {
  border-bottom-color: rgba(25,210,124,.7);
}

/* Modal */
.modal {
  position: fixed;
  inset: 0;
  display: grid;
  place-items: center;
  background: rgba(0,0,0,0.55);
  z-index: 1000;
  padding: 16px;
}
.modal-content {
  width: 100%;
  max-width: 420px;
  border-radius: 14px;
  border: 1px solid var(--color-border);
  background: var(--color-background);
  box-shadow: 0 14px 60px rgba(0,0,0,0.45);
  padding: 18px;
  text-align: center;
}
.modal-content h3 {
  margin-bottom: 6px;
  color: var(--color-heading);
  font-size: 18px;
}
.modal-content p {
  margin: 6px 0 12px;
  opacity: .9;
}
</style>
