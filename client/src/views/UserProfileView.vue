<script setup>
import { ref, onMounted, computed, watch } from "vue";
import { useRoute } from "vue-router";
import api from "../api/axios";

const route = useRoute();
const username = computed(() => route.params.username);

const profile = ref(null);
const viewer = ref({ username: "", role: "" });
const blogs = ref([]);
const selected = ref(null);

const isAdmin = computed(() => viewer.value.role === "admin");

const edit = ref({
  title: "",
  body: "",
  type: "",
  tags: "",
  category: "",
  status: "",
});

function normalizeBlog(b) {
  const deletedAtObj =
      b.deletedAt ?? b.DeletedAt ?? b.deleted_at ?? b.baseModel?.deletedAt ?? b.BaseModel?.DeletedAt ?? null;
  const deletedAt =
      (deletedAtObj && (deletedAtObj.Time || deletedAtObj.time || deletedAtObj)) || null;
  const isDeleted =
      !!deletedAtObj &&
      (deletedAtObj.Valid === true || !!deletedAtObj.Time || !!deletedAtObj.time || !!deletedAt);

  const isApprovedRaw = b.is_approved ?? b.isApproved ?? b.content?.is_approved ?? b.content?.isApproved ?? false;
  const statusRaw = b.status ?? b.content?.status ?? "";

  return {
    id: b.id ?? b.ID ?? b.baseModel?.id ?? b.BaseModel?.ID ?? null,
    title: b.title ?? b.content?.title ?? "",
    body: b.body ?? b.content?.body ?? "",
    username: b.username ?? b.content?.username ?? "",
    authorId: b.author_id ?? b.authorId ?? b.content?.author_id ?? b.content?.authorId ?? null,
    isApproved: isDeleted ? false : !!isApprovedRaw,
    status: isDeleted ? "deleted" : statusRaw,
    tags: b.tags ?? "",
    category: b.category ?? "",
    createdAt: b.createdAt ?? b.created_at ?? "",
    updatedAt: b.updatedAt ?? b.updated_at ?? "",
    deletedAt: deletedAt || "",
    raw: b,
  };
}

onMounted(load);

// route ile başka kullanıcıya geçişte otomatik reload
watch(
    () => route.params.username,
    async () => {
      profile.value = null;
      blogs.value = [];
      selected.value = null;
      await load();
    }
);

async function load() {
  // giriş yapan (viewer) -> /me
  try {
    const { data } = await api.get("/me");
    viewer.value = {
      username: data?.data?.username || "",
      role: data?.data?.role || "",
    };
  } catch {
    viewer.value = { username: "", role: "" };
  }

  // profil sahibi
  try {
    const { data } = await api.get(`/user/${encodeURIComponent(username.value)}`);
    profile.value = data?.data || null;
  } catch (e) {
    profile.value = null;
    console.error(e);
  }

  // bloglar
  try {
    const isSelf = viewer.value.username && viewer.value.username === username.value;
    const isViewerAdmin = viewer.value.role === "admin";

    // admin veya hesap sahibi → silinenler dahil
    const endpoint =
        isSelf || isViewerAdmin
            ? `/blogs-deleted/${encodeURIComponent(username.value)}`
            : `/blogs/${encodeURIComponent(username.value)}`;

    const res = await api.get(endpoint);
    blogs.value = (res.data?.data || []).map(normalizeBlog);
  } catch (e) {
    console.error("blogs load error", e);
  }
}

function selectBlog(b) {
  selected.value = b;
  edit.value = {
    title: b.title || "",
    body: b.body || "",
    type: b.type || "",
    tags: b.tags || "",
    category: b.category || "",
    status: b.status || "",
  };
}

// Admin aksiyonları
const canEdit = computed(
    () => isAdmin.value && selected.value && selected.value.status !== "deleted"
);

async function approveSelected() {
  if (!isAdmin.value || !selected.value) return;
  try {
    await api.put(`/blog/${encodeURIComponent(selected.value.title)}/approve`);
    selected.value.isApproved = true;

    // listeyi de güncelle
    const i = blogs.value.findIndex(
        (x) => (x.id ?? x.title) === (selected.value.id ?? selected.value.title)
    );
    if (i !== -1) blogs.value[i].isApproved = true;
  } catch (e) {
    alert(e?.response?.data?.error || "Onaylama başarısız");
  }
}

async function unapproveSelected() {
  if (!isAdmin.value || !selected.value) return;
  try {
    await api.put(`/blog/${encodeURIComponent(selected.value.title)}/unapprove`);
    selected.value.isApproved = false;
    const i = blogs.value.findIndex(
        (x) => (x.id ?? x.title) === (selected.value.id ?? selected.value.title)
    );
    if (i !== -1) blogs.value[i].isApproved = false;
  } catch (e) {
    alert(e?.response?.data?.error || "Onay kaldırma başarısız");
  }
}

async function restoreSelected() {
  if (!isAdmin.value || !selected.value) return;
  try {
    await api.put(`/blog/${encodeURIComponent(selected.value.title)}/restore`);
    // UI: geri yükleme → pending’e çek
    selected.value.status = "";
    selected.value.deletedAt = "";
    selected.value.isApproved = false;

    const i = blogs.value.findIndex(
        (x) => (x.id ?? x.title) === (selected.value.id ?? selected.value.title)
    );
    if (i !== -1) {
      blogs.value[i].status = "";
      blogs.value[i].deletedAt = "";
      blogs.value[i].isApproved = false;
    }
  } catch (e) {
    alert(e?.response?.data?.error || "Geri yükleme başarısız");
  }
}

async function updateSelected() {
  if (!canEdit.value || !selected.value) return;
  try {
    const oldTitle = selected.value.title;
    const payload = { ...edit.value };
    await api.put(`/blog/${encodeURIComponent(oldTitle)}`, payload);

    selected.value = { ...selected.value, ...payload, updatedAt: new Date().toISOString() };
    const i = blogs.value.findIndex(
        (x) => (x.id ?? x.title) === (selected.value.id ?? oldTitle)
    );
    if (i !== -1) blogs.value[i] = { ...blogs.value[i], ...payload };
    alert("Blog güncellendi.");
  } catch (e) {
    alert(e?.response?.data?.error || "Güncelleme başarısız");
  }
}

async function deleteSelected() {
  if (!isAdmin.value || !selected.value) return;
  if (!confirm(`Silinsin mi? (${selected.value.title})`)) return;
  try {
    await api.delete(`/blog/${encodeURIComponent(selected.value.title)}`);
    const i = blogs.value.findIndex(
        (x) => (x.id ?? x.title) === (selected.value.id ?? selected.value.title)
    );
    if (i !== -1) {
      blogs.value[i] = {
        ...blogs.value[i],
        status: "deleted",
        isApproved: false,
        deletedAt: new Date().toISOString(),
      };
      selected.value = { ...blogs.value[i] };
    }
    alert("Blog silindi (soft).");
  } catch (e) {
    alert(e?.response?.data?.error || "Silme başarısız");
  }
}
</script>

<template>
  <section class="page">
    <div class="header">
      <h1>Kullanıcı Profili</h1>
      <p><b>@{{ username }}</b> kullanıcısının profilini ve bloglarını görüntülüyorsunuz.</p>
    </div>

    <!-- Profil kartı -->
    <div class="card" v-if="profile">
      <div class="section-title">Profil</div>
      <div class="profile-grid">
        <div><b>Kullanıcı adı:</b> {{ profile.username }}</div>
        <div><b>Email:</b> {{ profile.email }}</div>
        <div><b>Rol:</b> {{ profile.role }}</div>
      </div>
    </div>
    <div v-else class="card muted">Profil yüklenemedi veya kullanıcı bulunamadı.</div>

    <div class="split">
      <!-- Sol: Liste -->
      <div class="card">
        <div class="section-title">Bloglar</div>

        <template v-if="blogs.length === 0">
          <p class="muted">Bu kullanıcıya ait blog bulunamadı.</p>
        </template>
        <ul v-else class="list">
          <li v-for="b in blogs" :key="b.id || b.title" class="item">
            <div class="item-left">
              <div class="line">
                <b class="title">{{ b.title }}</b>
                <span v-if="b.status === 'deleted'" class="badge badge-danger">silindi</span>
                <span v-else class="badge" :class="b.isApproved ? 'badge-success' : 'badge-warn'">
                  {{ b.isApproved ? 'approved' : 'pending' }}
                </span>
              </div>
              <div class="sub">Oluşturulma: {{ b.createdAt || '—' }}</div>
              <div class="preview">
                {{ (b.body || '').slice(0, 80) }}<span v-if="(b.body||'').length>80">…</span>
              </div>
            </div>

            <button class="btn btn-green" @click="selectBlog(b)">Seç</button>
          </li>
        </ul>
      </div>

      <!-- Sağ: Detay -->
      <div class="card">
        <div class="section-title">Detay</div>

        <div v-if="selected" class="detail">
          <div class="row">
            <div><b>Blog ID:</b> {{ selected.id ?? "—" }}</div>
            <div><b>Oluşturulma:</b> {{ selected.createdAt || "—" }}</div>
            <div><b>Güncellenme:</b> {{ selected.updatedAt || "—" }}</div>
            <div v-if="selected.deletedAt"><b>Silinme (soft):</b> {{ selected.deletedAt }}</div>
          </div>

          <div class="row"><b>Başlık:</b> {{ selected.title }}</div>

          <div class="row flex">
            <div class="line">
              <div><b>Yazar:</b> {{ selected.username }}</div>
              <span v-if="selected.status === 'deleted'" class="badge badge-danger">silindi</span>
              <span v-else class="badge" :class="selected.isApproved ? 'badge-success' : 'badge-warn'">
                {{ selected.isApproved ? 'approved' : 'pending' }}
              </span>
            </div>
          </div>

          <div class="grid">
            <div><b>AuthorID:</b> {{ selected.authorId ?? "—" }}</div>
            <div><b>Tip:</b> {{ selected.type || "—" }}</div>
            <div><b>Durum:</b> {{ selected.status || "—" }}</div>
            <div><b>Kategori:</b> {{ selected.category || "—" }}</div>
            <div class="span-2"><b>Etiketler:</b> {{ selected.tags || "—" }}</div>
          </div>

          <div class="row">
            <b>İçerik:</b>
            <div class="content">{{ selected.body }}</div>
          </div>

          <!-- Admin aksiyonları -->
          <div v-if="isAdmin" class="actions">
            <button
                v-if="selected.status === 'deleted'"
                class="btn btn-green"
                @click="restoreSelected"
            >
              Geri Yükle
            </button>

            <template v-else>
              <button
                  v-if="!selected.isApproved"
                  class="btn btn-green"
                  @click="approveSelected"
              >
                Onayla
              </button>
              <button
                  v-else
                  class="btn btn-red"
                  @click="unapproveSelected"
              >
                Onayı Kaldır
              </button>
            </template>

            <!-- Düzenleme / Sil (silinmemişse) -->
            <div v-if="canEdit" class="edit-box">
              <div class="grid">
                <div class="field"><label>Başlık</label><input v-model="edit.title" /></div>
                <div class="field"><label>Tip</label><input v-model="edit.type" /></div>
                <div class="field"><label>Etiketler</label><input v-model="edit.tags" /></div>
                <div class="field"><label>Kategori</label><input v-model="edit.category" /></div>
                <div class="field span-2"><label>Durum</label><input v-model="edit.status" /></div>
                <div class="field span-2"><label>İçerik</label><textarea v-model="edit.body" rows="6"></textarea></div>
              </div>

              <div class="row">
                <button class="btn btn-green" @click="updateSelected">Kaydet</button>
                <button class="btn btn-red" @click="deleteSelected">Sil</button>
              </div>
            </div>
          </div>
        </div>

        <p v-else class="muted">Listeden bir blog seçin.</p>
      </div>
    </div>
  </section>
</template>

<style scoped>
/* Sayfa iskeleti */
.page {
  max-width: 1000px;
  margin: 0 auto;
  display: grid;
  gap: 16px;
  padding: 8px 0 24px;
}
.header h1 {
  font-size: 28px;
  font-weight: 800;
  letter-spacing: .2px;
  color: var(--color-heading);
}
.header p { opacity: .8; margin-top: 2px; font-size: 14px; }

/* Kart */
.card {
  border: 1px solid var(--color-border);
  border-radius: 14px;
  padding: 16px;
  background:
      linear-gradient(0deg, rgba(255,255,255,0.02), rgba(255,255,255,0.02)) padding-box,
      radial-gradient(1200px 300px at 0% 0%, rgba(25,210,124,0.12), rgba(25,210,124,0) 60%),
      radial-gradient(1000px 300px at 100% 100%, rgba(25,210,124,0.10), rgba(25,210,124,0) 60%);
  backdrop-filter: blur(6px) saturate(120%);
  margin-bottom: 12px;
}
.section-title {
  font-weight: 800;
  color: var(--color-heading);
  margin-bottom: 10px;
}

/* İki sütun düzeni */
.split {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

/* Profil */
.profile-grid {
  display: grid;
  gap: 8px;
}

/* Liste */
.list { display: grid; gap: 8px; }
.item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  border: 1px solid var(--color-border);
  border-radius: 10px;
  padding: 10px;
  transition: background .15s ease, border-color .15s ease;
}
.item:hover {
  border-color: var(--color-border-hover);
  background: var(--color-background-soft);
}
.item-left { display: grid; gap: 6px; }
.line { display: flex; align-items: center; gap: 8px; }
.title { letter-spacing: .2px; }
.sub { opacity: .8; font-size: 13px; }
.preview { opacity: .9; margin-top: 2px; font-size: 13px; }

/* Detay */
.detail { display: grid; gap: 12px; }
.row { display: grid; gap: 4px; }
.flex { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}
.span-2 { grid-column: span 2; }
.content {
  margin-top: 6px;
  white-space: pre-wrap;
  border: 1px solid var(--color-border);
  border-radius: 10px;
  padding: 10px;
}

/* Form */
.edit-box { display: grid; gap: 12px; margin-top: 12px; border-top: 1px solid var(--color-border); padding-top: 12px; }
.field { display: grid; gap: 6px; }
.field label {
  font-size: 13px;
  font-weight: 700;
  color: var(--color-heading);
  letter-spacing: .2px;
}
.field input,
.field textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--color-border);
  border-radius: 10px;
  background: var(--color-background);
  color: var(--color-text);
  outline: none;
  transition: border-color .15s ease, box-shadow .15s ease;
}
.field input:focus,
.field textarea:focus {
  border-color: var(--color-border-hover);
  box-shadow: 0 0 0 2px hsla(160, 100%, 37%, 0.16);
}

/* Butonlar */
.btn {
  border: 1px solid transparent;
  border-radius: 8px;
  padding: 6px 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}
.btn-green {
  background: hsla(160, 100%, 37%, 1);
  border-color: hsla(160, 100%, 37%, 1);
  color: #fff;
}
.btn-green:hover { opacity: .92; box-shadow: 0 0 10px rgba(25,210,124,0.26); }
.btn-gray {
  background: rgba(255,255,255,0.05);
  border-color: rgba(255,255,255,0.2);
  color: #ddd;
}
.btn-gray:hover { border-color: rgba(255,255,255,0.35); background: rgba(255,255,255,0.08); }
.btn-red {
  background: rgba(255,0,0,.06);
  border-color: rgba(255,0,0,.35);
  color: #ff6e6e;
}
.btn-red:hover { background: rgba(255,0,0,.1); }

/* Rozetler */
.badge {
  font-size: 11px;
  line-height: 1;
  padding: 4px 6px;
  border-radius: 999px;
  border: 1px solid transparent;
  text-transform: lowercase;
  letter-spacing: .4px;
}
.badge-success {
  color: #0a7a3f;
  border-color: rgba(10,122,63,.25);
  background: rgba(10,122,63,.08);
}
.badge-warn {
  color: #946200;
  border-color: rgba(148,98,0,.25);
  background: rgba(148,98,0,.08);
}
.badge-danger {
  color: #b00;
  border-color: rgba(255,0,0,.25);
  background: rgba(255,0,0,.06);
}

.muted { opacity: .8; }

/* Mobil */
@media (max-width: 900px) {
  .split { grid-template-columns: 1fr; }
}
</style>