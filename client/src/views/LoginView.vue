<script setup>
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import api from "../api/axios";

const router = useRouter();

const identifier = ref("");
const password   = ref("");
const showPw     = ref(false);
const loading    = ref(false);
const error      = ref("");

// zaten login'liyse login sayfasına geleni ana sayfaya al
onMounted(() => {
  const token = localStorage.getItem("token");
  if (token) router.replace("/");
});

async function login() {
  if (loading.value) return;

  const id = identifier.value.trim();
  const pw = password.value;

  // basit frontend kontrolü
  if (!id || !pw) {
    error.value = "Lütfen kullanıcı/eposta ve şifre girin.";
    return;
  }

  loading.value = true;
  error.value = "";

  try {
    const { data } = await api.post("/login", {
      identifier: id,
      password: pw,
    });

    const d = data?.data || {};
    localStorage.setItem("token", d.token || "");
    localStorage.setItem("username", d.username || "");
    localStorage.setItem("email", d.email || "");
    localStorage.setItem("id", String(d.id || ""));
    localStorage.setItem("role", d.role || "");

    // header menüsünü tazele
    window.dispatchEvent(new Event("auth:changed"));

    // ana sayfaya
    router.push("/");
  } catch (e) {
    error.value = e?.response?.data?.error || "Giriş başarısız. Bilgileri kontrol edin.";
  } finally {
    loading.value = false;
  }
}

function onKeydown(e) {
  if (e.key === "Enter") login();
}
</script>

<template>
  <section class="login-wrap">
    <!-- Arka plan dekor -->
    <div class="bg">
      <div class="blob b1"></div>
      <div class="blob b2"></div>
    </div>

    <div class="card">
      <div class="brand">
        <img src="@/assets/logo-lognode.svg" alt="LogNode" class="brand-logo" />
        <div class="brand-text">
          <h1>LogNode</h1>
          <p>yaz, paylaş, keşfet</p>
        </div>
      </div>

      <div class="form">
        <label class="label">Email veya Kullanıcı Adı</label>
        <input
            class="input"
            v-model="identifier"
            placeholder="kullanici veya mail@ornek.com"
            :disabled="loading"
            @keydown="onKeydown"
        />

        <label class="label">Şifre</label>
        <div class="pw-row">
          <input
              class="input"
              :type="showPw ? 'text' : 'password'"
              v-model="password"
              placeholder="••••••••"
              :disabled="loading"
              @keydown="onKeydown"
          />
          <button
              class="pw-toggle"
              type="button"
              :disabled="loading"
              @click="showPw = !showPw"
              :aria-label="showPw ? 'Şifreyi gizle' : 'Şifreyi göster'"
          >
            {{ showPw ? "Gizle" : "Göster" }}
          </button>
        </div>

        <p v-if="error" class="error">{{ error }}</p>

        <button class="btn" :disabled="loading" @click="login">
          <span v-if="!loading">Giriş Yap</span>
          <span v-else class="spinner"></span>
        </button>

        <div class="sub-link">
          Hesabın yok mu?
          <router-link to="/register" class="link">
            Kayıt ol</router-link>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
/* Sayfa iskeleti */
.login-wrap {
  position: relative;
  min-height: calc(100vh - 120px); /* üst header çıktıktan sonra kalan alan */
  display: grid;
  place-items: center;
  padding: 24px 16px 40px;
  overflow: hidden;
}

/* Arka plan */
.bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
}
.blob {
  position: absolute;
  filter: blur(70px);
  opacity: 0.28;
  transform: translateZ(0);
}
.b1 {
  width: 360px; height: 360px; border-radius: 50%;
  background: radial-gradient(closest-side, #19d27c, transparent 70%);
  top: -80px; left: -60px;
}
.b2 {
  width: 420px; height: 420px; border-radius: 50%;
  background: radial-gradient(closest-side, #0bbf68, transparent 70%);
  bottom: -120px; right: -80px;
}

/* Kart */
.card {
  position: relative;
  width: 100%;
  max-width: 480px;
  background: color-mix(in oklab, var(--color-background) 86%, transparent);
  border: 1px solid var(--color-border);
  border-radius: 16px;
  padding: 22px 18px;
  box-shadow: 0 10px 30px rgba(0,0,0,.12);
  backdrop-filter: saturate(120%) blur(6px);
}

/* Marka bölümü */
.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 4px 4px 14px;
}
.brand-logo {
  width: 40px;
  height: 40px;
}
.brand-text h1 {
  margin: 0;
  font-size: 1.4rem;
  letter-spacing: .2px;
  color: var(--color-heading);
}
.brand-text p {
  margin: -4px 0 0;
  font-size: .92rem;
  opacity: .7;
}

/* Form */
.form {
  display: grid;
  gap: 10px;
}
.label {
  font-weight: 600;
  font-size: .95rem;
}
.input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--color-border);
  border-radius: 10px;
  background: var(--color-background);
  color: var(--color-text);
  outline: none;
  transition: border-color .2s, box-shadow .2s;
}
.input:focus {
  border-color: var(--color-border-hover);
  box-shadow: 0 0 0 2px hsla(160, 100%, 37%, 0.18) inset;
}

/* Şifre satırı */
.pw-row {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 8px;
  align-items: center;
}
.pw-toggle {
  border: 1px solid var(--color-border);
  background: var(--color-background-soft);
  color: var(--color-text);
  border-radius: 10px;
  padding: 9px 10px;
  cursor: pointer;
  transition: all .2s;
}
.pw-toggle:hover {
  border-color: var(--color-border-hover);
  background: var(--color-background-mute);
}

/* Hata */
.error {
  margin-top: 2px;
  color: #ff6b6b;
  font-size: .92rem;
}

/* Buton */
.btn {
  margin-top: 6px;
  width: 100%;
  padding: 11px 14px;
  border-radius: 12px;
  border: 1px solid #19d27c;
  background: linear-gradient(135deg, #19d27c, #0bbf68);
  color: #0b1510;
  font-weight: 700;
  letter-spacing: .2px;
  cursor: pointer;
  transition: transform .08s ease, filter .2s ease;
}
.btn:disabled {
  opacity: .7;
  cursor: not-allowed;
}
.btn:active {
  transform: translateY(1px);
}
.spinner {
  display: inline-block;
  width: 18px; height: 18px;
  border: 2px solid rgba(0,0,0,.25);
  border-top-color: rgba(0,0,0,.6);
  border-radius: 50%;
  animation: spin 0.85s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.link {
  color: #19d27c;
  text-decoration: none;
  border-bottom: 1px dashed rgba(25,210,124,.35);
  padding-bottom: 2px;
}
.link:hover {
  border-bottom-color: rgba(25,210,124,.7);
}

/* Alt link */
.sub-link {
  text-align: center;
  margin-top: 10px;
  display: inline-block;
  color: var(--color-text);
  opacity: .85;
  transition: opacity .2s;
}
.sub-link:hover { opacity: 1; }

/* Responsive ufak dokunuş */
@media (max-width: 520px) {
  .card { padding: 18px 14px; }
  .brand-text h1 { font-size: 1.25rem; }
}
</style>
