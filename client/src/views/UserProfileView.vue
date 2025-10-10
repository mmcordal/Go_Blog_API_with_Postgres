<!-- client/src/views/UserProfileView.vue -->
<script setup>
import { ref, onMounted, computed } from "vue";
import { useRoute } from "vue-router";
import api from "../api/axios";

const route = useRoute();
const username = computed(() => route.params.username);   // görüntülenecek kullanıcı

const profile = ref(null);
const viewer = ref({ username: "", role: "" });           // giriş yapan kişi
const blogs = ref([]);
const selected = ref(null);

// Admin mi?
const isAdmin = computed(() => viewer.value.role === "admin");

// Düzenleme formu (sadece admin)
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
      !!deletedAtObj && (deletedAtObj.Valid === true || !!deletedAtObj.Time || !!deletedAtObj.time || !!deletedAt);

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

async function load() {
  // giriş yapan (viewer)
  const myU = localStorage.getItem("username") || "";
  if (myU) {
    try {
      const { data } = await api.get(`/user/${encodeURIComponent(myU)}`);
      viewer.value = { username: data?.data?.username || "", role: data?.data?.role || "" };
    } catch {
      viewer.value = { username: myU, role: "" };
    }
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
    const endpoint = (isSelf || isViewerAdmin)
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

  // admin için düzenleme formunu doldur
  edit.value = {
    title: b.title || "",
    body: b.body || "",
    type: b.type || "",
    tags: b.tags || "",
    category: b.category || "",
    status: b.status || "",
  };
}

// --- Admin aksiyonları ---
const canEdit = computed(() => isAdmin.value && selected.value && selected.value.status !== "deleted");

async function approveSelected() {
  if (!isAdmin.value || !selected.value) return;
  try {
    await api.put(`/blog/${encodeURIComponent(selected.value.title)}/approve`);
    selected.value.isApproved = true;
    // listedeki öğeyi de güncelle
    const i = blogs.value.findIndex(x => (x.id ?? x.title) === (selected.value.id ?? selected.value.title));
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
    const i = blogs.value.findIndex(x => (x.id ?? x.title) === (selected.value.id ?? selected.value.title));
    if (i !== -1) blogs.value[i].isApproved = false;
  } catch (e) {
    alert(e?.response?.data?.error || "Onay kaldırma başarısız");
  }
}

async function restoreSelected() {
  if (!isAdmin.value || !selected.value) return;
  try {
    await api.put(`/blog/${encodeURIComponent(selected.value.title)}/restore`);
    // geri yüklenince statü/flagleri düzelt
    selected.value.status = "";
    selected.value.deletedAt = "";
    // istersen pending’e çekebilirsin:
    // selected.value.isApproved = false;

    const i = blogs.value.findIndex(x => (x.id ?? x.title) === (selected.value.id ?? selected.value.title));
    if (i !== -1) {
      blogs.value[i].status = "";
      blogs.value[i].deletedAt = "";
      // blogs.value[i].isApproved = false;
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

    // Detay ve listeyi güncelle
    selected.value = { ...selected.value, ...payload, updatedAt: new Date().toISOString() };
    const i = blogs.value.findIndex(x => (x.id ?? x.title) === (selected.value.id ?? oldTitle));
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
    // Soft delete → listede işaretli kalsın
    const i = blogs.value.findIndex(x => (x.id ?? x.title) === (selected.value.id ?? selected.value.title));
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
  <section style="max-width:1000px;margin:auto;display:grid;gap:16px">
    <h1>Kullanıcı Profili</h1>

    <div v-if="profile" style="display:grid;gap:8px;border:1px solid #eee;padding:12px;border-radius:8px">
      <div><b>Kullanıcı adı:</b> {{ profile.username }}</div>
      <div><b>Email:</b> {{ profile.email }}</div>
      <div><b>Rol:</b> {{ profile.role }}</div>
    </div>
    <p v-else>Profil yüklenemedi veya kullanıcı bulunamadı.</p>

    <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px">
      <!-- Liste -->
      <div>
        <h2>Bloglar</h2>
        <ul style="display:grid;gap:6px">
          <li
              v-for="b in blogs"
              :key="b.id || b.title"
              style="display:flex;justify-content:space-between;align-items:center;border:1px solid #eee;padding:8px;border-radius:8px"
          >
            <div>
              <b>{{ b.title }}</b>
              <span v-if="b.status === 'deleted'" style="margin-left:8px;color:#b00;font-weight:bold">[silindi]</span>
              <span v-else style="margin-left:8px;font-size:12px;opacity:.8">[{{ b.isApproved ? 'approved' : 'pending' }}]</span>
              <div style="opacity:.8;margin-top:4px">
                {{ (b.body || '').slice(0, 80) }}<span v-if="(b.body||'').length>80">…</span>
              </div>
            </div>
            <button @click="selectBlog(b)">Seç</button>
          </li>
        </ul>
      </div>

      <!-- Detay -->
      <div>
        <h3>Detay</h3>
        <div v-if="selected" style="display:grid;gap:12px;border:1px solid #eee;padding:12px;border-radius:8px">
          <div><b>Blog ID:</b> {{ selected.id ?? "—" }}</div>
          <div><b>Oluşturulma:</b> {{ selected.createdAt || "—" }}</div>
          <div><b>Güncellenme:</b> {{ selected.updatedAt || "—" }}</div>
          <div v-if="selected.deletedAt"><b>Silinme (soft):</b> {{ selected.deletedAt }}</div>

          <div><b>Başlık:</b> {{ selected.title }}</div>
          <div><b>Yazar Username:</b> {{ selected.username }}</div>
          <div><b>AuthorID:</b> {{ selected.authorId ?? "—" }}</div>
          <div><b>Tip:</b> {{ selected.type || "—" }}</div>
          <div><b>Durum:</b> {{ selected.status || "—" }}</div>
          <div><b>Onay:</b> {{ selected.isApproved ? "Evet" : "Hayır" }}</div>
          <div><b>Etiketler:</b> {{ selected.tags || "—" }}</div>
          <div><b>Kategori:</b> {{ selected.category || "—" }}</div>
          <div><b>İçerik:</b></div>
          <div style="white-space:pre-wrap">{{ selected.body }}</div>

          <!-- ADMIN AKSİYONLARI -->
          <div v-if="isAdmin" class="admin-actions">
            <div class="row">
              <!-- Silinmişse sadece Geri Yükle -->
              <button v-if="selected.status === 'deleted'" @click="restoreSelected">Geri Yükle</button>

              <!-- Silinmemişse onay/ret -->
              <template v-else>
                <button v-if="!selected.isApproved" @click="approveSelected">Onayla</button>
                <button v-else @click="unapproveSelected">Onayı Kaldır</button>
              </template>
            </div>

            <!-- Düzenleme / Sil (silinmemişse) -->
            <div v-if="canEdit" class="edit-box">
              <h4>Güncelle</h4>

              <div class="field">
                <label for="t">Başlık</label>
                <input id="t" v-model="edit.title" />
              </div>

              <div class="field">
                <label for="b">İçerik</label>
                <textarea id="b" v-model="edit.body" rows="6"></textarea>
              </div>

              <div class="field">
                <label for="ty">Tip</label>
                <input id="ty" v-model="edit.type" />
              </div>

              <div class="field">
                <label for="tg">Etiketler</label>
                <input id="tg" v-model="edit.tags" />
              </div>

              <div class="field">
                <label for="ct">Kategori</label>
                <input id="ct" v-model="edit.category" />
              </div>

              <div class="field">
                <label for="st">Durum</label>
                <input id="st" v-model="edit.status" />
              </div>

              <div class="row">
                <button @click="updateSelected">Kaydet</button>
                <button class="danger" @click="deleteSelected">Sil</button>
              </div>
            </div>
          </div>
        </div>

        <p v-else>Listeden bir blog seçin.</p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.admin-actions {
  display: grid;
  gap: 12px;
  margin-top: 8px;
  border-top: 1px solid #eee;
  padding-top: 12px;
}
.row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.edit-box {
  display: grid;
  gap: 10px;
}
.field {
  display: grid;
  gap: 6px;
}
.field label {
  font-weight: 600;
}
.field input, .field textarea {
  width: 100%;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 8px;
  box-sizing: border-box;
}
button.danger {
  color: #b00;
}
</style>