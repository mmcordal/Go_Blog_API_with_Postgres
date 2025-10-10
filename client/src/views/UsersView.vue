<script setup>
import { ref, computed, onMounted, watch } from "vue";
import { useRoute } from "vue-router";
import api from "../api/axios";
import SearchBar from "../components/SearchBar.vue";

const route = useRoute();

const mode = ref("user");
const q = ref("");
const users = ref([]);
const selected = ref(null);

const me = ref({ username: "", role: "reader" });
const isAdmin = computed(() => me.value.role === "admin");

// giriş yapan kullanıcı bilgisi
onMounted(async () => {
  const myUsername = localStorage.getItem("username");
  if (myUsername) {
    try {
      const { data } = await api.get(`/user/${encodeURIComponent(myUsername)}`);
      me.value = {
        username: data?.data?.username,
        role: data?.data?.role,
        email: data?.data?.email,
      };
    } catch (e) {
      console.error("rol/me alınamadı:", e);
    }
  }

  // URL ?username=... varsa aramayı doldur ve otomatik seç
  prefillFromQuery();
});

// URL değişirse tekrar dene
watch(() => route.query.username, () => {
  prefillFromQuery();
});

async function prefillFromQuery() {
  const uname = route.query?.username;
  if (!uname) return;

  mode.value = "user";
  q.value = String(uname);
  await onSearch({ q: q.value, mode: "user" });

  const found = users.value.find(u =>
      (u.username || "").toLowerCase() === String(uname).toLowerCase()
  );
  if (found) selectUser(found);
}

async function onSearch({ q: query, mode: m }) {
  q.value = query;
  mode.value = m;
  if (m !== "user") return;
  if (!query) { users.value = []; return; }

  try {
    const { data } = await api.get("/users", { params: { search: query, limit: 10 } });
    users.value = data?.data || [];
  } catch (e) {
    console.error(e);
  }
}

function selectUser(u) {
  selected.value = { ...u, newUsername: u.username, newEmail: u.email, newPassword: "" };
}

// SADECE admin düzenleyebilir
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
    selected.value.username = selected.value.newUsername;
    selected.value.email = selected.value.newEmail;
    selected.value.newPassword = "";
  } catch (e) {
    alert(e?.response?.data?.error || "Güncelleme başarısız");
  }
}

async function deleteUser() {
  if (!selected.value) return;
  if (!confirm(`Silinsin mi? (${selected.value.username})`)) return;
  try {
    await api.delete(`/user/${encodeURIComponent(selected.value.username)}`);
    alert("Kullanıcı silindi (soft delete)");
    users.value = users.value.filter(u => (u.id || u.username) !== (selected.value.id || selected.value.username));
    selected.value = null;
  } catch (e) {
    alert(e?.response?.data?.error || "Silme başarısız");
  }
}
</script>

<template>
  <section style="max-width:900px;margin:auto;display:grid;gap:16px">
    <h1>Kullanıcılar</h1>

    <SearchBar v-model="q" :mode="mode" @update:mode="v => mode=v" @search="onSearch" />

    <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px">
      <!-- Sol: Liste -->
      <div>
        <h3>Sonuçlar</h3>
        <ul style="display:grid;gap:6px">
          <li
              v-for="u in users"
              :key="u.id || u.username"
              style="display:flex;justify-content:space-between;align-items:center;border:1px solid #eee;padding:8px;border-radius:8px"
          >
            <div>
              <b>{{ u.username }}</b>
              <div style="opacity:.7">{{ u.email }}</div>
              <div style="opacity:.7">Rol: {{ u.role }}</div>
            </div>

            <div style="display:flex; gap:8px; align-items:center">
              <button @click="selectUser(u)">Seç</button>
              <!-- Profili Aç: ileride /u/:username rotasına da bağlayabiliriz -->
              <router-link :to="`/u/${encodeURIComponent(u.username)}`">
                <button style="border:1px solid #333;border-radius:6px;padding:6px 10px;cursor:pointer">
                  Profili Aç
                </button>
              </router-link>
            </div>
          </li>
        </ul>
      </div>

      <!-- Sağ: Detay -->
      <div>
        <h3>{{ isAdmin ? 'Detay / Güncelle / Sil' : 'Detay' }}</h3>

        <div
            v-if="selected"
            style="display:grid;gap:8px;border:1px solid #eee;padding:12px;border-radius:8px"
        >
          <div><b>ID:</b> {{ selected.id }}</div>
          <div><b>Kullanıcı adı:</b> {{ selected.username }}</div>
          <div><b>Email:</b> {{ selected.email }}</div>
          <div><b>Rol:</b> {{ selected.role }}</div>

          <!-- Düzenleme alanı: SADECE admin -->
          <div v-if="canEditSelected()" class="edit-box">
            <div class="field">
              <label for="nu">Yeni kullanıcı adı</label>
              <input id="nu" v-model="selected.newUsername" placeholder="Kullanıcı adı" />
            </div>

            <div class="field">
              <label for="ne">Yeni email</label>
              <input id="ne" v-model="selected.newEmail" placeholder="Email" />
            </div>

            <div class="field">
              <label for="np">Yeni şifre (opsiyonel)</label>
              <input id="np" v-model="selected.newPassword" placeholder="Şifre" type="password" />
            </div>

            <div class="actions">
              <button @click="updateUser">Güncelle</button>
              <button class="danger" @click="deleteUser">Sil</button>
            </div>
          </div>

          <div v-else style="opacity:.7;margin-top:8px">
            Bu bölüm yalnızca <b>admin</b> tarafından düzenlenebilir.
          </div>
        </div>

        <p v-else>Listeden bir kullanıcı seçin.</p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.edit-box {
  display: grid;
  gap: 12px;
}

.field {
  display: grid;
  gap: 6px;
}

.field label {
  font-weight: 600;
}

.field input {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 8px;
  box-sizing: border-box;
}

.actions {
  display: flex;
  gap: 8px;
  margin-top: 4px;
}

.actions .danger {
  color: #b00;
}
</style>
