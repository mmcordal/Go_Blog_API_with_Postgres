<script setup>
import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import api from "../api/axios";

const router = useRouter();

const page = ref(1);
const perPage = ref(10);

const q = ref("");
const allBlogs = ref([]);
const selected = ref(null);

// admin kontrolü
const me = ref({ role: "" });
const isAdmin = computed(() => me.value.role === "admin");

// buton loading state
const loading = ref({ approve: false, unapprove: false, restore: false });

/** normalize — MyAccount / BlogsAll ile aynı şablon */
function normalize(b){
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
    authorId,
    isApproved: isDeleted ? false : !!isApprovedRaw,
    status: isDeleted ? "deleted" : statusRaw,
    tags: b.tags ?? "",
    category: b.category ?? "",
    createdAt: b.createdAt ?? b.created_at ?? "",
    updatedAt: b.updatedAt ?? b.updated_at ?? "",
    deletedAt: deletedAt || "",
  };
}

onMounted(async () => {
  try {
    // rol kontrolü (admin -> include_deleted)
    const u = localStorage.getItem("username");
    let includeDeleted = false;
    if (u) {
      try {
        const { data } = await api.get(`/me`);
        me.value.role = data?.data?.role || "";
        includeDeleted = me.value.role === "admin";
      } catch {}
    }

    const { data } = await api.get("/blogs", {
      params: includeDeleted ? { include_deleted: true } : {},
    });
    allBlogs.value = (data?.data || []).map(normalize);
  } catch (e) {
    console.error("blogs load error", e);
  }
});

// q boşsa sonuç gösterme
const results = computed(() => {
  const s = q.value.trim().toLowerCase();
  if (!s) return [];
  return allBlogs.value.filter(b => (b.title || "").toLowerCase().startsWith(s));
});

const paginatedResults = computed(() => {
const start = (page.value - 1) * perPage.value;
return results.value.slice(start, start + perPage.value);
});

function selectBlog(b){ selected.value = b; }

// --- Admin aksiyonları ---
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
    <div class="header">
      <h1>Blog Ara</h1>
      <p>Başlığa göre arayın, detayları görüntüleyin. Adminler silinenleri görebilir ve onaylayabilir.</p>
    </div>

    <!-- Arama kutusu -->
    <div class="card">
      <div class="search-row">
        <input
            v-model.trim="q"
            placeholder="Başlıkla ara (en az 1 harf)…"
            class="input"
        />
        <router-link to="/blogs/all">
          <button class="primary-btn">
            Tüm Blogları Gör
          </button>
        </router-link>
      </div>
    </div>

    <div class="split">
      <!-- Liste -->
      <div class="card">
        <div class="section-title">Sonuçlar</div>
        <p v-if="q && results.length === 0" class="muted">Eşleşen blog bulunamadı.</p>

        <ul v-else-if="q.trim()" class="list">
          <li v-for="b in paginatedResults" :key="b.id || b.title" class="item">
            <div class="item-left">
              <div class="line">
                <b class="title">{{ b.title }}</b>
                <span v-if="b.status === 'deleted'" class="badge badge-danger">silindi</span>
                <span v-else class="badge" :class="b.isApproved ? 'badge-success' : 'badge-warn'">
                  {{ b.isApproved ? 'approved' : 'pending' }}
                </span>
              </div>
              <div class="sub">Yazar: {{ b.username }}</div>
              <div class="preview">
                {{ (b.body || '').slice(0, 80) }}<span v-if="(b.body||'').length>80">…</span>
              </div>
            </div>

            <button
                class="btn btn-green" style="border:1px solid #333;border-radius:6px;padding:6px 10px;cursor:pointer"
                @click="selectBlog(b)"
            >Seç
            </button>
          </li>
        </ul>

        <p v-else class="muted">Arama yapmaya başlayın…</p>

        <div class="pagination" v-if="results.length > perPage">
          <button :disabled="page===1" @click="page--">‹ Önceki</button>
          <span>Sayfa {{ page }} / {{ Math.ceil(results.length / perPage) }}</span>
          <button :disabled="page >= Math.ceil(results.length / perPage)" @click="page++">Sonraki ›</button>
        </div>
      </div>

      <!-- Detay -->
      <div class="card">
        <div class="section-title">Detay</div>

        <div v-if="selected" class="detail">
          <div class="row">
            <div><b>Blog ID:</b> {{ selected.id ?? '—' }}</div>
            <div><b>Oluşturulma:</b> {{ selected.createdAt }}</div>
            <div><b>Güncellenme:</b> {{ selected.updatedAt }}</div>
            <div v-if="selected.deletedAt"><b>Silinme (soft):</b> {{ selected.deletedAt }}</div>
          </div>

          <div class="row"><b>Başlık:</b> {{ selected.title }}</div>

          <div class="row flex">
            <div><b>Yazar:</b> {{ selected.username }}</div>
            <router-link :to="`/u/${encodeURIComponent(selected.username)}`">
              <button class="btn btn-gray" style="padding:6px 10px;border:1px solid #888;border-radius:6px;cursor:pointer">
                Yazar Profilini Aç
              </button>
            </router-link>
          </div>

          <div class="grid">
            <div><b>AuthorID:</b> {{ selected.authorId ?? '—' }}</div>
            <div><b>Durum:</b> {{ selected.status || '—' }}</div>
            <div><b>Onay:</b> {{ selected.isApproved ? 'Evet' : 'Hayır' }}</div>
            <div><b>Kategori:</b> {{ selected.category || '—' }}</div>
            <div class="span-2"><b>Etiketler:</b> {{ selected.tags || '—' }}</div>
          </div>

          <div class="row">
            <b>İçerik:</b>
            <div class="content">{{ selected.body }}</div>
          </div>

          <!-- Admin aksiyonları -->
          <div v-if="isAdmin" class="actions">
            <!-- Silinmişse sadece geri yükle -->
            <button
                class="btn btn-green"
                v-if="selected.status === 'deleted'"
                style="border:1px solid #333;border-radius:6px;padding:6px 10px;cursor:pointer"
                @click="restoreSelected"
                :disabled="loading.restore"
            >
              {{ loading.restore ? 'Geri Yükleniyor…' : 'Geri Yükle' }}
            </button>

            <!-- Silinmemişse onay/ret -->
            <template v-else>
              <button
                  class="btn btn-green"
                  v-if="!selected.isApproved"
                  style="border:1px solid #333;border-radius:6px;padding:6px 10px;cursor:pointer"
                  @click="approveSelected"
                  :disabled="loading.approve"
              >
                {{ loading.approve ? 'Onaylanıyor…' : 'Onayla' }}
              </button>

              <button
                  class="btn btn-red"
                  v-else
                  style="border:1px solid #333;border-radius:6px;padding:6px 10px;cursor:pointer"
                  @click="unapproveSelected"
                  :disabled="loading.unapprove"
              >
                {{ loading.unapprove ? 'Kaldırılıyor…' : 'Onayı Kaldır' }}
              </button>
            </template>
          </div>
        </div>

        <p v-else class="muted">Arama yapın ve listeden bir blog seçin.</p>
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
}

.section-title {
  font-weight: 800;
  color: var(--color-heading);
  margin-bottom: 10px;
}

/* Arama satırı */
.search-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.input {
  flex: 1;
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

/* İki sütun */
.split {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
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
.preview { opacity: .85; margin-top: 2px; font-size: 13px; }

/* Detay */
.detail { display: grid; gap: 10px; }
.row { display: grid; gap: 4px; }
.flex { display: flex; gap: 10px; align-items: center; flex-wrap: wrap; }
.grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px 14px;
}
.span-2 { grid-column: span 2; }
.content {
  white-space: pre-wrap;
  border: 1px solid var(--color-border);
  border-radius: 10px;
  padding: 10px;
  background: var(--color-background);
}

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

.primary-btn {
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  border: 1px solid hsla(160, 100%, 37%, 1);
  background: hsla(160, 100%, 37%, 1); /* LogNode yeşili */
  color: #fff;
  transition: transform .05s ease, box-shadow .15s ease, opacity .15s ease;
}
.primary-btn:hover {
  opacity: .95;
  box-shadow: 0 0 0 2px hsla(160, 100%, 37%, 0.18) inset;
}
.primary-btn:active {
  transform: translateY(1px);
}
/* LogNode temalı genel buton stilleri */
.btn {
  border: 1px solid transparent;
  border-radius: 8px;
  padding: 6px 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

/* Yeşil (Onay / Seç / Oluştur) */
.btn-green {
  background: hsla(160, 100%, 37%, 1);
  border-color: hsla(160, 100%, 37%, 1);
  color: #fff;
}
.btn-green:hover {
  opacity: .9;
  box-shadow: 0 0 10px rgba(25,210,124,0.3);
}

/* Gri (Nötr işlemler / Geri dön / İptal) */
.btn-gray {
  background: rgba(255,255,255,0.05);
  border-color: rgba(255,255,255,0.2);
  color: #ddd;
}
.btn-gray:hover {
  border-color: rgba(255,255,255,0.35);
  background: rgba(255,255,255,0.08);
}

/* Kırmızı (Sil / Reddet / Kaldır) */
.btn-red {
  background: rgba(176,0,0,0.9);
  border-color: rgba(176,0,0,1);
  color: #fff;
}
.btn-red:hover {
  background: rgba(176,0,0,1);
  box-shadow: 0 0 8px rgba(255,0,0,0.3);
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  margin-top: 12px;
}
.pagination button {
  padding: 6px 10px;
  border-radius: 6px;
  border: 1px solid var(--color-border);
  background: var(--color-background-soft);
  color: var(--color-text);
  cursor: pointer;
}
.pagination button:hover { background: var(--color-background); }
.pagination button:disabled { opacity: .5; cursor: default; }
</style>
