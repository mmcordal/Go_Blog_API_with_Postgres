<script setup>
import { ref, computed, onMounted, watch } from "vue";
import api from "../api/axios";

const allBlogs = ref([]);
const selected = ref(null);
const q = ref("");

// admin kontrolü
const me = ref({ role: "" });
const isAdmin = computed(() => me.value.role === "admin");

// aksiyon loading’leri
const loading = ref({ approve: false, unapprove: false, restore: false });

// pagination
const page = ref(1);
const perPage = ref(10);
const perPageOptions = [6, 12, 24, 48];
const PERPAGE_KEY = "blogsAll:perPage";

/** normalize (diğer sayfalarla uyumlu) */
function normalizeBlog(b) {
  const deletedAtObj =
      b.deletedAt ?? b.DeletedAt ?? b.deleted_at ?? b.baseModel?.deletedAt ?? b.BaseModel?.DeletedAt ?? null;
  const deletedAt = (deletedAtObj && (deletedAtObj.Time || deletedAtObj.time || deletedAtObj)) || null;
  const isDeleted = !!deletedAtObj && (deletedAtObj.Valid === true || !!deletedAtObj.Time || !!deletedAtObj.time || !!deletedAt);

  const authorId =
      b.author_id ?? b.authorId ?? b.AuthorID ?? b.content?.author_id ?? b.content?.authorId ?? b.Content?.AuthorID ?? null;

  const isApprovedRaw = b.is_approved ?? b.isApproved ?? b.content?.is_approved ?? b.content?.isApproved ?? false;
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
    // perPage'i hatırla
    const saved = Number(localStorage.getItem(PERPAGE_KEY) || "");
    if (perPageOptions.includes(saved)) perPage.value = saved;

    // rolü çek → admin ise include_deleted
    let includeDeleted = false;
    try {
      const { data } = await api.get("/me");
      me.value.role = data?.data?.role || "";
      includeDeleted = me.value.role === "admin";
    } catch (_) {}

    const { data } = await api.get("/blogs", {
      params: includeDeleted ? { include_deleted: true } : {},
    });
    allBlogs.value = (data?.data || []).map(normalizeBlog);
    page.value = 1;
  } catch (e) {
    console.error("blogs load error", e);
  }
}

/* ------- Arama + Sayfalama ------- */
const filtered = computed(() => {
  const s = q.value.trim().toLowerCase();
  if (!s) return allBlogs.value;
  return allBlogs.value.filter(b => (b.title || "").toLowerCase().includes(s));
});

const totalPages = computed(() => Math.max(1, Math.ceil((filtered.value.length || 0) / perPage.value)));
const paginated = computed(() => {
  const start = (page.value - 1) * perPage.value;
  return filtered.value.slice(start, start + perPage.value);
});

// arama/perPage değişince sayfayı başa al + perPage'i kaydet
watch(q, () => { page.value = 1; });
watch(perPage, (n) => {
  localStorage.setItem(PERPAGE_KEY, String(n));
  page.value = 1;
});
// liste kısalınca sayfayı clamp et
watch([filtered, perPage], () => {
  if (page.value > totalPages.value) page.value = totalPages.value;
});

// pager helpers
function first(){ page.value = 1; }
function prev(){ page.value = Math.max(1, page.value - 1); }
function next(){ page.value = Math.min(totalPages.value, page.value + 1); }
function last(){ page.value = totalPages.value; }

function selectBlog(b) { selected.value = b; }

/* ------- Admin aksiyonları ------- */
async function approveSelected() {
  if (!selected.value || loading.value.approve) return;
  loading.value.approve = true;
  try {
    await api.put(`/blog/${encodeURIComponent(selected.value.title)}/approve`);
    selected.value.isApproved = true;
    const i = allBlogs.value.findIndex(x => (x.id ?? x.title) === (selected.value.id ?? selected.value.title));
    if (i !== -1) allBlogs.value[i].isApproved = true;
  } catch (e) {
    alert(e?.response?.data?.error || "Onaylama başarısız");
  } finally {
    loading.value.approve = false;
  }
}

async function unapproveSelected() {
  if (!selected.value || loading.value.unapprove) return;
  loading.value.unapprove = true;
  try {
    await api.put(`/blog/${encodeURIComponent(selected.value.title)}/unapprove`);
    selected.value.isApproved = false;
    const i = allBlogs.value.findIndex(x => (x.id ?? x.title) === (selected.value.id ?? selected.value.title));
    if (i !== -1) allBlogs.value[i].isApproved = false;
  } catch (e) {
    alert(e?.response?.data?.error || "Onayı kaldırma başarısız");
  } finally {
    loading.value.unapprove = false;
  }
}

async function restoreSelected() {
  if (!selected.value || loading.value.restore) return;
  loading.value.restore = true;
  try {
    await api.put(`/blog/${encodeURIComponent(selected.value.title)}/restore`);
    selected.value.status = "";
    selected.value.deletedAt = "";
    const i = allBlogs.value.findIndex(x => (x.id ?? x.title) === (selected.value.id ?? selected.value.title));
    if (i !== -1) {
      allBlogs.value[i].status = "";
      allBlogs.value[i].deletedAt = "";
    }
  } catch (e) {
    alert(e?.response?.data?.error || "Geri yükleme başarısız");
  } finally {
    loading.value.restore = false;
  }
}
</script>

<template>
  <section class="page">
    <div class="page-top">
      <h1>Tüm Bloglar</h1>
      <router-link to="/blogs" class="back-link">← Aramaya dön</router-link>
    </div>

    <!-- Arama / PerPage -->
    <div class="card">
      <div class="search-row">
        <input v-model="q" placeholder="Listede filtrele…" class="input" />
        <div class="perpage">
          <label>Sayfa başına</label>
          <select v-model.number="perPage" class="select">
            <option v-for="n in perPageOptions" :key="n" :value="n">{{ n }}</option>
          </select>
        </div>

        <div class="pager">
          <button class="btn btn-gray" @click="first" :disabled="page===1">⏮</button>
          <button class="btn btn-gray" @click="prev"  :disabled="page===1">←</button>
          <span class="page-info">Sayfa {{ page }} / {{ totalPages }}</span>
          <button class="btn btn-gray" @click="next"  :disabled="page===totalPages">→</button>
          <button class="btn btn-gray" @click="last"  :disabled="page===totalPages">⏭</button>
        </div>
      </div>
    </div>

    <div class="grid2">
      <!-- Liste -->
      <div>
        <div class="list-head">
          <h3>Liste</h3>
          <div class="muted small">Toplam: {{ filtered.length }}</div>
        </div>

        <ul class="list">
          <li
              v-for="b in paginated"
              :key="b.id || b.title"
              class="list-card"
          >
            <div class="list-main">
              <div class="line">
                <b class="title">{{ b.title }}</b>
                <span v-if="b.status === 'deleted'" class="badge badge-danger">silindi</span>
                <span v-else class="badge" :class="b.isApproved ? 'badge-success' : 'badge-warn'">
                  {{ b.isApproved ? 'approved' : 'pending' }}
                </span>
              </div>

              <div class="list-body">
                {{ (b.body || '').slice(0, 80) }}<span v-if="(b.body||'').length>80">…</span>
              </div>
              <div class="list-meta">Yazar: {{ b.username }}</div>
            </div>

            <button class="btn btn-green" style="border:1px solid #333;border-radius:6px;padding:6px 10px;cursor:pointer" @click="selectBlog(b)">
              Seç
            </button>
          </li>
        </ul>

        <!-- Alt sayfalama -->
        <div class="pagination" v-if="filtered.length > perPage">
          <button :disabled="page===1" @click="first">⏮</button>
          <button :disabled="page===1" @click="prev">‹ Önceki</button>
          <span>Sayfa {{ page }} / {{ totalPages }}</span>
          <button :disabled="page>=totalPages" @click="next">Sonraki ›</button>
          <button :disabled="page>=totalPages" @click="last">⏭</button>
        </div>
      </div>

      <!-- Detay -->
      <div>
        <h3>Detay</h3>
        <div v-if="selected" class="detail-card">
          <div class="kv"><b>Blog ID:</b> <span>{{ selected.id ?? '—' }}</span></div>
          <div class="kv"><b>Oluşturulma:</b> <span>{{ selected.createdAt || '—' }}</span></div>
          <div class="kv"><b>Güncellenme:</b> <span>{{ selected.updatedAt || '—' }}</span></div>
          <div class="kv" v-if="selected.deletedAt"><b>Silinme (soft):</b> <span>{{ selected.deletedAt }}</span></div>

          <div class="kv"><b>Başlık:</b> <span>{{ selected.title }}</span></div>

          <div class="row-wrap">
            <div class="kv"><b>Yazar:</b> <span>{{ selected.username }}</span></div>
            <router-link :to="`/u/${encodeURIComponent(selected.username)}`">
              <button class="btn btn-gray" style="padding:6px 10px;border:1px solid #888;border-radius:6px;cursor:pointer">
                Yazar Profilini Aç
              </button>
            </router-link>
          </div>

          <div class="kv"><b>AuthorID:</b> <span>{{ selected.authorId ?? '—' }}</span></div>
          <div class="kv"><b>Tip:</b> <span>{{ selected.type || '—' }}</span></div>
          <div class="kv"><b>Durum:</b> <span>{{ selected.status || '—' }}</span></div>
          <div class="kv"><b>Onay:</b> <span>{{ selected.isApproved ? 'Evet' : 'Hayır' }}</span></div>
          <div class="kv"><b>Etiketler:</b> <span>{{ selected.tags || '—' }}</span></div>
          <div class="kv"><b>Kategori:</b> <span>{{ selected.category || '—' }}</span></div>

          <div class="kv-col">
            <b>İçerik:</b>
            <div class="content-pre">{{ selected.body }}</div>
          </div>

          <!-- Admin aksiyonları -->
          <div v-if="isAdmin" class="actions">
            <button
                class="btn btn-green"
                v-if="selected.status === 'deleted'"
                style="border:1px solid #333;border-radius:6px;padding:6px 10px;cursor:pointer"
                @click="restoreSelected"
                :disabled="loading.restore"
            >
              {{ loading.restore ? "Geri Yükleniyor…" : "Geri Yükle" }}
            </button>

            <template v-else>
              <button
                  class="btn btn-green"
                  v-if="!selected.isApproved"
                  style="border:1px solid #333;border-radius:6px;padding:6px 10px;cursor:pointer"
                  @click="approveSelected"
                  :disabled="loading.approve"
              >
                {{ loading.approve ? "Onaylanıyor…" : "Onayla" }}
              </button>
              <button
                  class="btn btn-red"
                  v-else
                  style="border:1px solid #333;border-radius:6px;padding:6px 10px;cursor:pointer"
                  @click="unapproveSelected"
                  :disabled="loading.unapprove"
              >
                {{ loading.unapprove ? "Kaldırılıyor…" : "Onayı Kaldır" }}
              </button>
            </template>
          </div>
        </div>

        <p v-else class="muted">Listeden bir blog seçin.</p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.page {
  max-width: 1000px;
  margin: auto;
  display: grid;
  gap: 16px;
}

.page-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.back-link {
  text-decoration: none;
  font-size: 14px;
  opacity: .9;
  border: 1px solid var(--color-border);
  padding: 6px 10px;
  border-radius: 8px;
  transition: background .15s ease, box-shadow .15s ease;
}
.back-link:hover {
  background: var(--color-background-soft);
  box-shadow: 0 0 0 2px hsla(160, 100%, 37%, .12) inset;
}

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
}
.search-row {
  display: flex; align-items: center; gap: 12px; flex-wrap: wrap;
}
.input {
  flex: 1;
  min-width: 220px;
  padding: 10px 12px;
  border: 1px solid var(--color-border);
  border-radius: 10px;
  background: var(--color-background);
  color: var(--color-text);
  outline: none;
  transition: border-color .15s ease, box-shadow .15s ease;
}
.input:focus {
  border-color: var(--color-border-hover);
  box-shadow: 0 0 0 2px hsla(160, 100%, 37%, 0.16);
}

/* per page + pager */
.perpage { display: flex; align-items: center; gap: 8px; }
.perpage label { font-size: 12px; opacity: .85; }
.select {
  padding: 8px 10px;
  border: 1px solid var(--color-border);
  border-radius: 10px;
  background: var(--color-background);
  color: var(--color-text);
}
.pager { display: inline-flex; gap: 6px; align-items: center; }
.page-info { min-width: 110px; text-align: center; }

/* 2 kolon grid */
.grid2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
@media (max-width: 900px) { .grid2 { grid-template-columns: 1fr; } }

/* Liste başlığı */
.list-head {
  display: flex; align-items: baseline; justify-content: space-between; margin: 6px 0 8px;
}
.small { font-size: 12px; }

/* Liste kartları */
.list { display: grid; gap: 6px; padding: 0; margin: 0; list-style: none; }
.list-card {
  display: flex; justify-content: space-between; align-items: center;
  border: 1px solid var(--color-border); padding: 10px; border-radius: 12px;
  transition: background .15s ease, border-color .15s ease, transform .05s ease;
}
.list-card:hover {
  border-color: var(--color-border-hover);
  background:
      linear-gradient(0deg, rgba(255,255,255,0.02), rgba(255,255,255,0.02)) padding-box,
      radial-gradient(600px 200px at 0% 0%, rgba(25,210,124,0.10), rgba(25,210,124,0) 60%);
}
.list-main { display: grid; gap: 6px; }
.list-body { opacity: .85; }
.list-meta { opacity: .6; font-size: 12px; }

/* Detay kartı */
.detail-card {
  display: grid; gap: 10px;
  border: 1px solid var(--color-border);
  padding: 12px; border-radius: 12px;
  background:
      linear-gradient(0deg, rgba(255,255,255,0.02), rgba(255,255,255,0.02)) padding-box,
      radial-gradient(700px 260px at 100% 0%, rgba(25,210,124,0.10), rgba(25,210,124,0) 60%);
}
.kv { display: grid; grid-template-columns: 160px 1fr; gap: 8px; align-items: baseline; }
.kv-col { display: grid; gap: 6px; }
.content-pre { white-space: pre-wrap; border: 1px solid var(--color-border); border-radius: 8px; padding: 10px; }

.row-wrap { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }

.actions { margin-top: 12px; display: flex; gap: 8px; flex-wrap: wrap; }
.muted { opacity: .7; }

.line { display: flex; align-items: center; gap: 8px; }
.title { letter-spacing: .2px; }

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
.badge-success { color: #0a7a3f; border-color: rgba(10,122,63,.25); background: rgba(10,122,63,.08); }
.badge-warn    { color: #946200; border-color: rgba(148,98,0,.25); background: rgba(148,98,0,.08); }
.badge-danger  { color: #b00; border-color: rgba(255,0,0,.25); background: rgba(255,0,0,.06); }

/* Pagination (alt) */
.pagination {
  display: flex; justify-content: center; align-items: center; gap: 10px; margin-top: 12px;
}
.pagination button {
  padding: 6px 10px; border-radius: 6px; cursor: pointer;
  border: 1px solid var(--color-border); background: var(--color-background-soft); color: var(--color-text);
  transition: background .15s ease;
}
.pagination button:hover { background: var(--color-background); }
.pagination button:disabled { opacity: .5; cursor: default; }

/* LogNode temalı butonlar */
.btn { border: 1px solid transparent; border-radius: 8px; padding: 6px 12px; font-weight: 600; cursor: pointer; transition: all 0.2s ease; }
.btn-green { background: hsla(160, 100%, 37%, 1); border-color: hsla(160, 100%, 37%, 1); color: #fff; }
.btn-green:hover { opacity: .9; box-shadow: 0 0 10px rgba(25,210,124,0.3); }
.btn-gray { background: rgba(255,255,255,0.05); border-color: rgba(255,255,255,0.2); color: #ddd; }
.btn-gray:hover { border-color: rgba(255,255,255,0.35); background: rgba(255,255,255,0.08); }
.btn-red { background: rgba(176,0,0,0.9); border-color: rgba(176,0,0,1); color: #fff; }
.btn-red:hover { background: rgba(176,0,0,1); box-shadow: 0 0 8px rgba(255,0,0,0.3); }
</style>
