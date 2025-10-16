<script setup>
import { ref, computed, onMounted, watch } from "vue";
import { useRoute } from "vue-router";
import api from "../api/axios";

const route = useRoute();
const mode = ref("user");
const q = ref("");
const users = ref([]);
const selected = ref(null);
const me = ref({ username: "", role: "reader" });
const isAdmin = computed(() => me.value.role === "admin");
const loading = ref(false);
const noResults = ref(false);
let searchTimeout = null;

// giriş yapan kullanıcı bilgisi
onMounted(async () => {
  const myUsername = localStorage.getItem("username");
  if (myUsername) {
    try {
      const { data } = await api.get(`/me`);
      me.value = {
        username: data?.data?.username,
        role: data?.data?.role,
        email: data?.data?.email,
      };
    } catch (e) {
      console.error("rol/me alınamadı:", e);
    }
  }
  prefillFromQuery();
});

watch(() => route.query.username, prefillFromQuery);

// --- normalize: silinmiş kullanıcıyı doğru hesapla
function normalizeUser(u) {
  const raw =
      u.deletedAt ??
      u.DeletedAt ??
      u.deleted_at ??
      u.BaseModel?.DeletedAt ??
      u.baseModel?.deletedAt ??
      null;

  let deletedAt = "";
  let isDeleted = false;

  if (raw && typeof raw === "object") {
    const t = raw.Time || raw.time || "";
    const v = raw.Valid === true || raw.valid === true;
    if (v && t) {
      deletedAt = t;
      isDeleted = true;
    }
  } else if (typeof raw === "string") {
    const d = new Date(raw);
    const ok = !isNaN(d.getTime()) && d.getFullYear() > 1971;
    if (ok) {
      deletedAt = raw;
      isDeleted = true;
    }
  }

  return {
    id: u.id ?? u.ID ?? null,
    username: u.username ?? "",
    email: u.email ?? "",
    role: u.role ?? "",
    deletedAt,
    isDeleted,
    raw: u,
  };
}

async function prefillFromQuery() {
  const uname = route.query?.username;
  if (!uname) return;
  q.value = String(uname);
  await fetchUsers();

  // ✅ arama sonrası otomatik seç
  const found = users.value.find(
      (u) => (u.username || "").toLowerCase() === String(uname).toLowerCase()
  );
  if (found) selectUser(found);
}

function debouncedSearch() {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(fetchUsers, 400);
}

async function fetchUsers() {
  if (!q.value.trim()) {
    users.value = [];
    noResults.value = false;
    return;
  }

  try {
    loading.value = true;
    const params = {
      search: q.value.trim(),
      limit: 10,
      ...(isAdmin.value ? { include_deleted: true } : {}),
    };
    const { data } = await api.get("/users", { params });
    const list = (data?.data || []).map(normalizeUser);
    users.value = list;
    noResults.value = list.length === 0;
  } catch (e) {
    console.error(e);
  } finally {
    loading.value = false;
  }
}

function selectUser(u) {
  selected.value = {
    ...u,
    newUsername: u.username,
    newEmail: u.email,
    newPassword: "",
  };
}

function canEditSelected() {
  return isAdmin.value;
}

async function updateUser() {
  if (!selected.value) return;
  try {
    const body = {
      username: selected.value.newUsername,
      email: selected.value.newEmail,
    };
    if (selected.value.newPassword) body.password = selected.value.newPassword;

    await api.put(`/user/${encodeURIComponent(selected.value.username)}`, body);
    alert("Kullanıcı güncellendi");

    // ✅ UI senkronu: selected + list item
    selected.value.username = selected.value.newUsername;
    selected.value.email = selected.value.newEmail;
    selected.value.newPassword = "";

    const i = users.value.findIndex(
        (u) => (u.id || u.username) === (selected.value.id || selected.value.username)
    );
    if (i !== -1) {
      users.value[i] = {
        ...users.value[i],
        username: selected.value.username,
        email: selected.value.email,
      };
    }
  } catch (e) {
    alert(e?.response?.data?.error || "Güncelleme başarısız");
  }
}

async function deleteUser() {
  if (!selected.value) return;
  if (!confirm(`Silinsin mi? (${selected.value.username})`)) return;
  try {
    await api.delete(`/user/${encodeURIComponent(selected.value.username)}`);
    alert("Kullanıcı silindi");
    users.value = users.value.filter(
        (u) => u.username !== selected.value.username
    );
    selected.value = null;
  } catch (e) {
    alert(e?.response?.data?.error || "Silme başarısız");
  }
}

async function restoreUser() {
  if (!selected.value?.isDeleted) return;
  try {
    await api.put(`/user/${encodeURIComponent(selected.value.username)}/restore`);
    alert("Kullanıcı geri yüklendi.");

    // ✅ UI senkronu
    selected.value.isDeleted = false;
    selected.value.deletedAt = "";

    const i = users.value.findIndex(
        (u) => (u.id || u.username) === (selected.value.id || selected.value.username)
    );
    if (i !== -1) {
      users.value[i] = { ...users.value[i], isDeleted: false, deletedAt: "" };
    }
  } catch (e) {
    alert(e?.response?.data?.error || "Geri yükleme başarısız");
  }
}
</script>

<template>
  <section class="page">
    <div class="header">
      <h1>Kullanıcılar</h1>
      <p>Kullanıcıları arayın, profillerini görüntüleyin. Adminler silinmiş hesapları da görebilir.</p>
    </div>

    <div class="card">
      <div class="search-row">
        <input
            v-model="q"
            placeholder="Kullanıcı adıyla ara…"
            class="input"
            @input="debouncedSearch"
        />
      </div>
    </div>

    <div class="split">
      <div class="card">
        <div class="section-title">Sonuçlar</div>

        <p v-if="loading" class="muted">Yükleniyor...</p>
        <p v-else-if="noResults" class="muted">Eşleşen kullanıcı bulunamadı.</p>
        <p v-else-if="!q.trim()" class="muted">Arama yapmaya başlayın…</p>

        <ul v-else class="list">
          <li v-for="u in users" :key="u.id || u.username" class="item">
            <div class="item-left">
              <div class="userline">
                <b class="uname">{{ u.username }}</b>
                <span v-if="u.isDeleted" class="badge badge-danger">silindi</span>
              </div>
              <div class="sub">{{ u.email }}</div>
              <div class="sub">Rol: {{ u.role }}</div>
            </div>

            <div class="item-actions">
              <button @click="selectUser(u)" class="btn mini-btn">Seç</button>
              <router-link :to="`/u/${encodeURIComponent(u.username)}`">
                <button class="btn btn-gray">Profili Aç</button>
              </router-link>
            </div>
          </li>
        </ul>
      </div>

      <div class="card">
        <div class="section-title">Detay</div>

        <div v-if="selected" class="detail">
          <div class="row">
            <div><b>Kullanıcı ID:</b> {{ selected.id ?? '—' }}</div>
            <div v-if="selected.deletedAt"><b>Silinme (soft):</b> {{ selected.deletedAt }}</div>
          </div>

          <div class="row line">
            <b class="title">{{ selected.username }}</b>
            <span v-if="selected.isDeleted" class="badge badge-danger">silindi</span>
          </div>

          <div class="row flex">
            <div><b>Email:</b> {{ selected.email }}</div>
          </div>

          <div class="grid">
            <div><b>Rol:</b> {{ selected.role }}</div>
          </div>

          <template v-if="canEditSelected()">
            <div v-if="selected.isDeleted" class="actions">
              <button class="mini-btn" @click="restoreUser">Geri Yükle</button>
            </div>

            <div v-else class="edit-box">
              <div class="grid">
                <div class="field">
                  <label for="nu">Yeni kullanıcı adı</label>
                  <input id="nu" v-model="selected.newUsername" placeholder="Kullanıcı adı" />
                </div>

                <div class="field">
                  <label for="ne">Yeni email</label>
                  <input id="ne" v-model="selected.newEmail" placeholder="Email" />
                </div>

                <div class="field span-2">
                  <label for="np">Yeni şifre (opsiyonel)</label>
                  <input id="np" v-model="selected.newPassword" type="password" placeholder="Şifre" />
                </div>
              </div>

              <div class="actions">
                <button class="mini-btn" @click="updateUser">Güncelle</button>
                <button class="mini-btn danger" @click="deleteUser">Sil</button>
              </div>
            </div>
          </template>

          <div v-else class="muted">
            Bu bölüm yalnızca <b>admin</b> tarafından düzenlenebilir.
          </div>
        </div>

        <p v-else class="muted">Listeden bir kullanıcı seçin.</p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.page { max-width: 1000px; margin: auto; display: grid; gap: 16px; padding: 20px 0; }
.header h1 { font-size: 28px; font-weight: 800; color: var(--color-heading); }
.card {
  border: 1px solid var(--color-border);
  border-radius: 14px;
  padding: 16px;
  background: linear-gradient(0deg, rgba(255,255,255,.02), rgba(255,255,255,.02)) padding-box,
  radial-gradient(800px 300px at 0% 0%, rgba(25,210,124,.12), transparent 60%);
  backdrop-filter: blur(6px) saturate(120%);
}
.search-row { display: flex; align-items: center; gap: 12px; }
.input {
  flex: 1; padding: 10px 12px; border: 1px solid var(--color-border);
  border-radius: 10px; background: var(--color-background);
  color: var(--color-text); outline: none;
}
.input:focus {
  border-color: var(--color-border-hover);
  box-shadow: 0 0 0 2px hsla(160, 100%, 37%, 0.16);
}
.split { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
@media (max-width: 900px){ .split { grid-template-columns: 1fr; } }

.list { display: grid; gap: 8px; }
.item {
  display: flex; justify-content: space-between; align-items: center;
  border: 1px solid var(--color-border);
  border-radius: 10px; padding: 10px;
  transition: border-color .15s ease, background .15s ease;
}
.item:hover { border-color: var(--color-border-hover); background: var(--color-background-soft); }
.item-left { display: grid; gap: 4px; }
.userline { display: flex; align-items: center; gap: 8px; }
.badge { font-size: 11px; padding: 4px 6px; border-radius: 999px; text-transform: lowercase; }
.badge-danger { color: #b00; background: rgba(255,0,0,.06); border: 1px solid rgba(255,0,0,.25); }

.btn {
  border: 1px solid transparent;
  border-radius: 8px;
  padding: 6px 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-gray {
  background: rgba(255,255,255,0.05);
  border-color: rgba(255,255,255,0.2);
  color: #ddd;
}
.btn-gray:hover {
  border-color: rgba(255,255,255,0.35);
  background: rgba(255,255,255,0.08);
}

.mini-btn {
  border: 1px solid #19d27c;
  color: #fff;
  background: #19d27c;
  border-radius: 6px;
  padding: 6px 10px;
  cursor: pointer;
  transition: opacity .15s ease;
}
.mini-btn:hover { opacity: .9; }
.danger { background: #b00; border-color: #b00; }
.muted { opacity: .8; font-size: 14px; }

/* --- Siyah temalı input alanları --- */
.field input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--color-border);
  border-radius: 10px;
  background: #111;
  color: #f2f2f2;
  outline: none;
  transition: border-color .15s ease, box-shadow .15s ease, background .15s ease;
}
.field input::placeholder { color: rgba(255,255,255,0.5); }
.field input:focus {
  border-color: #19d27c;
  background: #0b0b0b;
  box-shadow: 0 0 0 2px rgba(25,210,124,0.2);
}

/* Admin düzenleme kutusu */
.edit-box {
  display: grid;
  gap: 14px;
  margin-top: 12px;
  border-top: 1px solid var(--color-border);
  padding-top: 14px;
  background: rgba(255,255,255,0.02);
  border-radius: 12px;
}
</style>
