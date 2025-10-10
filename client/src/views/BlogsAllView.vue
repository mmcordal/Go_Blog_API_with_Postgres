<script setup>
import { ref, computed, onMounted } from "vue";
import api from "../api/axios";

const allBlogs = ref([]);
const selected = ref(null);
const q = ref("");

// admin kontrolü için
const me = ref({ role: "" });
const isAdmin = computed(() => me.value.role === "admin");

/** MyAccount ile senkron normalize */
function normalizeBlog(b) {
  const deletedAtObj =
      b.deletedAt ??
      b.DeletedAt ??
      b.deleted_at ??
      b.baseModel?.deletedAt ??
      b.BaseModel?.DeletedAt ??
      null;

  const deletedAt =
      (deletedAtObj && (deletedAtObj.Time || deletedAtObj.time || deletedAtObj)) || null;

  const isDeleted =
      !!deletedAtObj &&
      (deletedAtObj.Valid === true || !!deletedAtObj.Time || !!deletedAtObj.time || !!deletedAt);

  const authorId =
      b.author_id ??
      b.authorId ??
      b.AuthorID ??
      b.content?.author_id ??
      b.content?.authorId ??
      b.Content?.AuthorID ??
      null;

  const isApprovedRaw =
      b.is_approved ?? b.isApproved ?? b.content?.is_approved ?? b.content?.isApproved ?? false;

  const statusRaw = b.status ?? b.content?.status ?? "";

  return {
    id: b.id ?? b.ID ?? b.baseModel?.id ?? b.BaseModel?.ID ?? null,
    title: b.title ?? b.content?.title ?? "",
    body: b.body ?? b.content?.body ?? "",
    username: b.username ?? b.content?.username ?? "",
    type: b.type ?? b.content?.type ?? "",
    authorId,
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

onMounted(loadAll);

async function loadAll() {
  try {
    // rolü çek
    const myUsername = localStorage.getItem("username");
    let includeDeleted = false;
    if (myUsername) {
      try {
        const { data } = await api.get(`/user/${encodeURIComponent(myUsername)}`);
        me.value.role = data?.data?.role || "";
        includeDeleted = me.value.role === "admin";
      } catch (_) {}
    }

    // admin ise include_deleted paramıyla iste
    const { data } = await api.get("/blogs", {
      params: includeDeleted ? { include_deleted: true } : {},
    });

    allBlogs.value = (data?.data || []).map(normalizeBlog);
  } catch (e) {
    console.error("blogs load error", e);
  }
}

const results = computed(() => {
  const s = q.value.trim().toLowerCase();
  if (!s) return allBlogs.value;
  return allBlogs.value.filter(b => (b.title || "").toLowerCase().includes(s));
});

function selectBlog(b) { selected.value = b; }

// --- Admin aksiyonları ---
async function approveSelected() {
  if (!selected.value) return;
  await api.put(`/blog/${encodeURIComponent(selected.value.title)}/approve`);
  selected.value.isApproved = true;
  const i = allBlogs.value.findIndex(x => (x.id ?? x.title) === (selected.value.id ?? selected.value.title));
  if (i !== -1) allBlogs.value[i].isApproved = true;
}

async function unapproveSelected() {
  if (!selected.value) return;
  await api.put(`/blog/${encodeURIComponent(selected.value.title)}/unapprove`);
  selected.value.isApproved = false;
  const i = allBlogs.value.findIndex(x => (x.id ?? x.title) === (selected.value.id ?? selected.value.title));
  if (i !== -1) allBlogs.value[i].isApproved = false;
}

async function restoreSelected() {
  if (!selected.value) return;
  await api.put(`/blog/${encodeURIComponent(selected.value.title)}/restore`);
  selected.value.status = "";
  selected.value.deletedAt = "";
  // istersen pending’e çekebilirsin:
  // selected.value.isApproved = false;

  const i = allBlogs.value.findIndex(x => (x.id ?? x.title) === (selected.value.id ?? selected.value.title));
  if (i !== -1) {
    allBlogs.value[i].status = "";
    allBlogs.value[i].deletedAt = "";
    // allBlogs.value[i].isApproved = false;
  }
}
</script>

<template>
  <section style="max-width:1000px;margin:auto;display:grid;gap:16px">
    <div style="display:flex;justify-content:space-between;align-items:center">
      <h1>Tüm Bloglar</h1>
      <router-link to="/blogs">← Aramaya dön</router-link>
    </div>

    <input
        v-model="q"
        placeholder="Listede filtrele…"
        style="padding:8px;border:1px solid #ddd;border-radius:8px"
    />

    <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px">
      <!-- Liste -->
      <div>
        <h3>Liste</h3>
        <ul style="display:grid;gap:6px">
          <li
              v-for="b in results"
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
              <div style="opacity:.6; font-size:12px; margin-top:4px">Yazar: {{ b.username }}</div>
            </div>
            <button @click="selectBlog(b)">Seç</button>
          </li>
        </ul>
      </div>

      <!-- Detay -->
      <div>
        <h3>Detay</h3>
        <div v-if="selected" style="display:grid;gap:8px;border:1px solid #eee;padding:12px;border-radius:8px">
          <div><b>Blog ID:</b> {{ selected.id ?? '—' }}</div>
          <div><b>Oluşturulma:</b> {{ selected.createdAt || '—' }}</div>
          <div><b>Güncellenme:</b> {{ selected.updatedAt || '—' }}</div>

          <div v-if="selected.deletedAt"><b>Silinme (soft):</b> {{ selected.deletedAt }}</div>

          <div><b>Başlık:</b> {{ selected.title }}</div>
          <div style="display:flex;gap:8px;align-items:center;flex-wrap:wrap">
            <div><b>Yazar:</b> {{ selected.username }}</div>

            <!-- Yazar Profili butonu -->
            <router-link :to="`/u/${encodeURIComponent(selected.username)}`">
              <button style="padding:6px 10px;border:1px solid #888;border-radius:6px;cursor:pointer">Yazar Profilini Aç</button>
            </router-link>
          </div>


          <div><b>AuthorID:</b> {{ selected.authorId ?? '—' }}</div>
          <div><b>Tip:</b> {{ selected.type || '—' }}</div>
          <div><b>Durum:</b> {{ selected.status || '—' }}</div>
          <div><b>Onay:</b> {{ selected.isApproved ? 'Evet' : 'Hayır' }}</div>
          <div><b>Etiketler:</b> {{ selected.tags || '—' }}</div>
          <div><b>Kategori:</b> {{ selected.category || '—' }}</div>
          <div><b>İçerik:</b></div>
          <div style="white-space:pre-wrap">{{ selected.body }}</div>

          <!-- Admin aksiyonları -->
          <div v-if="isAdmin" style="margin-top:12px;display:flex;gap:8px;flex-wrap:wrap">
            <button v-if="selected.status === 'deleted'" @click="restoreSelected">Geri Yükle</button>

            <template v-else>
              <button v-if="!selected.isApproved" @click="approveSelected">Onayla</button>
              <button v-else @click="unapproveSelected">Onayı Kaldır</button>
            </template>
          </div>
        </div>
        <p v-else>Listeden bir blog seçin.</p>
      </div>
    </div>
  </section>
</template>
