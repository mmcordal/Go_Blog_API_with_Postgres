<script setup>
import { ref, computed, onMounted } from "vue";
import api from "../api/axios";

const allBlogs = ref([]);
const selected = ref(null);
const q = ref("");

function normalizeBlog(b) {
  const blogId =
      b.id ?? b.ID ?? b.blogId ?? b.BlogID ?? b.baseModel?.id ?? b.BaseModel?.ID;

  const authorId =
      b.authorId ??
      b.AuthorID ??
      b.content?.authorId ??
      b.Content?.AuthorID ??
      b.content?.id;

  return {
    id: blogId || null,
    title: b.title ?? b.content?.title ?? "",
    body: b.body ?? b.content?.body ?? "",
    username: b.username ?? b.content?.username ?? "",
    type: b.type ?? b.content?.type ?? "",
    isApproved:
        b.is_approved ??
        b.isApproved ??
        b.content?.is_approved ??
        b.content?.isApproved ??
        false,
    status: b.status ?? b.content?.status ?? "",
    tags: b.tags ?? "",
    category: b.category ?? "",
    createdAt: b.createdAt ?? b.created_at ?? "",
    updatedAt: b.updatedAt ?? b.updated_at ?? "",
    authorId: authorId ?? null,
    raw: b,
  };
}

onMounted(loadAll);

async function loadAll() {
  try {
    const { data } = await api.get("/blogs");
    allBlogs.value = (data?.data || []).map(normalizeBlog);
  } catch (e) {
    console.error("blogs load error", e);
  }
}

const results = computed(() => {
  if (!q.value) return allBlogs.value;
  const s = q.value.toLowerCase();
  return allBlogs.value.filter(b => (b.title || "").toLowerCase().startsWith(s));
});

function selectBlog(b) {
  selected.value = b;
}
</script>

<template>
  <section style="max-width:1000px;margin:auto;display:grid;gap:16px">
    <h1>Bloglar</h1>

    <input
        v-model="q"
        placeholder="Başlıkla ara (en az 1 harf)…"
        style="padding:8px;border:1px solid #ddd;border-radius:8px"
    />

    <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px">
      <!-- Liste -->
      <div>
        <h3>Sonuçlar</h3>
        <ul style="display:grid;gap:6px">
          <li
              v-for="b in results"
              :key="(b.id || b.title)"
              style="display:flex;justify-content:space-between;align-items:center;border:1px solid #eee;padding:8px;border-radius:8px"
          >
            <div>
              <b>{{ b.title }}</b>
              <div style="opacity:.7">Yazar: {{ b.username }}</div>
              <div style="opacity:.8; margin-top:4px">
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
        <div v-if="selected" style="display:grid;gap:8px;border:1px solid #eee;padding:12px;border-radius:8px">
          <div><b>Blog ID:</b> {{ selected.id ?? '—' }}</div>
          <div><b>Oluşturulma:</b> {{ selected.createdAt || '—' }}</div>
          <div><b>Güncellenme:</b> {{ selected.updatedAt || '—' }}</div>
          <div><b>Başlık:</b> {{ selected.title }}</div>
          <div><b>Yazar Username:</b> {{ selected.username }}</div>
          <div><b>AuthorID:</b> {{ selected.authorId ?? '—' }}</div>
          <div><b>Tip:</b> {{ selected.type || '—' }}</div>
          <div><b>Durum:</b> {{ selected.status || '—' }}</div>
          <div><b>Onay:</b> {{ selected.isApproved ? 'Evet' : 'Hayır' }}</div>
          <div><b>Etiketler:</b> {{ selected.tags || '—' }}</div>
          <div><b>Kategori:</b> {{ selected.category || '—' }}</div>
          <div><b>İçerik:</b></div>
          <div style="white-space:pre-wrap">{{ selected.body }}</div>
        </div>
        <p v-else>Listeden bir blog seçin.</p>
      </div>
    </div>
  </section>
</template>
