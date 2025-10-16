<script setup>
import { ref, onMounted, computed, watch } from "vue";
import api from "../api/axios";

/* --- Me / Admin kontrolü --- */
const me = ref({ role: "" });
const isAdmin = computed(() => (me.value.role || "").toLowerCase() === "admin");

/* --- Liste ve seçim --- */
const all = ref([]);
const pending = computed(() =>
    all.value.filter(b => !b.isApproved && b.status !== "deleted")
);
const selected = ref(null);

/* --- UI durumları --- */
const loading = ref(false); // liste yükleniyor
const actionLoading = ref({ approve: false, unapprove: false }); // buton kilidi
const fetchError = ref(""); // liste hatası

/* --- Sayfalama --- */
const perPageOptions = [10, 20, 50];
const perPage = ref(1);
const page = ref(1);

const pendingCount = computed(() => pending.value.length);
const totalPages = computed(() =>
    Math.max(1, Math.ceil(pendingCount.value / perPage.value || 1))
);
const paginated = computed(() => {
  const start = (page.value - 1) * perPage.value;
  return pending.value.slice(start, start + perPage.value);
});
function goFirst(){ page.value = 1; }
function goPrev(){ page.value = Math.max(1, page.value - 1); }
function goNext(){ page.value = Math.min(totalPages.value, page.value + 1); }
function goLast(){ page.value = totalPages.value; }

/* değişimlerde sayfayı resetle / clamp et */
watch([perPage, pending], () => {
  page.value = 1;
});
watch(page, (p) => {
  if (p > totalPages.value) page.value = totalPages.value;
});

/* --- Normalize: diğer sayfalarla senkron --- */
function normalize(b = {}) {
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

  const isApprovedRaw =
      b.is_approved ?? b.isApproved ?? b.content?.is_approved ?? b.content?.isApproved ?? false;

  const statusRaw = b.status ?? b.content?.status ?? "";

  return {
    id: b.id ?? b.ID ?? b.baseModel?.id ?? b.BaseModel?.ID ?? null,
    title: b.title ?? b.content?.title ?? "",
    body: b.body ?? b.content?.body ?? "",
    username: b.username ?? b.content?.username ?? "",
    authorId:
        b.author_id ?? b.authorId ?? b.content?.author_id ?? b.content?.authorId ?? null,
    isApproved: isDeleted ? false : !!isApprovedRaw,
    status: isDeleted ? "deleted" : statusRaw,
    tags: b.tags ?? "",
    category: b.category ?? "",
    createdAt: b.createdAt ?? b.created_at ?? "",
    updatedAt: b.updatedAt ?? b.updated_at ?? "",
    deletedAt: deletedAt || "",
  };
}

/* --- Yükleme --- */
async function load() {
  fetchError.value = "";
  loading.value = true;
  selected.value = null;

  try {
    try {
      const { data } = await api.get("/me");
      me.value.role = data?.data?.role || "";
    } catch {
      me.value.role = "";
    }
    if (!isAdmin.value) {
      all.value = [];
      return;
    }

    const { data } = await api.get("/blogs", { params: { include_deleted: true } });
    all.value = (data?.data || []).map(normalize);
    page.value = 1; // yeni veri -> başa dön
  } catch (e) {
    fetchError.value = e?.response?.data?.error || e?.message || "Liste alınamadı.";
    all.value = [];
  } finally {
    loading.value = false;
  }
}

onMounted(load);

function selectBlog(b) { selected.value = b; }

/* --- Admin aksiyonları --- */
async function approveSelected() {
  if (!selected.value || actionLoading.value.approve) return;
  actionLoading.value.approve = true;
  try {
    await api.put(`/blog/${encodeURIComponent(selected.value.title)}/approve`);
    selected.value.isApproved = true;
    all.value = all.value.map(x =>
        (x.id ?? x.title) === (selected.value.id ?? selected.value.title)
            ? { ...x, isApproved: true }
            : x
    );
  } catch (e) {
    alert(e?.response?.data?.error || "Onay başarısız");
  } finally {
    actionLoading.value.approve = false;
  }
}

async function unapproveSelected() {
  if (!selected.value || actionLoading.value.unapprove) return;
  actionLoading.value.unapprove = true;
  try {
    await api.put(`/blog/${encodeURIComponent(selected.value.title)}/unapprove`);
    selected.value.isApproved = false;
    all.value = all.value.map(x =>
        (x.id ?? x.title) === (selected.value.id ?? selected.value.title)
            ? { ...x, isApproved: false }
            : x
    );
  } catch (e) {
    alert(e?.response?.data?.error || "Onayı kaldırma başarısız");
  } finally {
    actionLoading.value.unapprove = false;
  }
}
</script>

<template>
  <section class="page">
    <div class="header">
      <h1>Onay Bekleyen Bloglar</h1>
      <p>Yalnızca <b>admin</b> bu sayfayı görebilir ve blog onay durumlarını yönetebilir.</p>
    </div>

    <div v-if="!isAdmin" class="muted card">
      Bu sayfayı sadece adminler görebilir.
    </div>

    <div v-else class="flow">
      <div class="card toolbar-card">
        <div class="toolbar">
          <div class="left">
            <b>Bekleyen: {{ pendingCount }}</b>
          </div>
          <div class="right">
            <label class="label" for="perPageSel">Sayfa başı</label>
            <select id="perPageSel" v-model.number="perPage" class="select" :disabled="loading">
              <option v-for="n in perPageOptions" :key="n" :value="n">{{ n }}</option>
            </select>

            <div class="pager">
              <button class="btn btn-gray" @click="goFirst" :disabled="loading || page===1">⏮</button>
              <button class="btn btn-gray" @click="goPrev"  :disabled="loading || page===1">←</button>
              <span class="page-info">{{ page }} / {{ totalPages }}</span>
              <button class="btn btn-gray" @click="goNext"  :disabled="loading || page===totalPages">→</button>
              <button class="btn btn-gray" @click="goLast"  :disabled="loading || page===totalPages">⏭</button>
            </div>

            <button class="btn btn-gray" @click="load" :disabled="loading">
              {{ loading ? 'Yükleniyor…' : 'Yenile' }}
            </button>
          </div>
        </div>
        <p v-if="fetchError" class="muted" style="margin-top:8px">Hata: {{ fetchError }}</p>
      </div>

      <div class="split">
        <!-- Liste -->
        <div class="card">
          <div class="section-title">Bekleyenler <span class="count">({{ pendingCount }})</span></div>

          <p v-if="loading" class="muted">Yükleniyor…</p>
          <p v-else-if="pendingCount === 0" class="muted">Şu an onay bekleyen blog yok.</p>

          <ul v-else class="list">
            <li v-for="b in paginated" :key="b.id || b.title" class="item">
              <div class="item-left">
                <div class="line">
                  <b class="title">{{ b.title }}</b>
                  <span class="badge badge-warn">pending</span>
                </div>
                <div class="sub">Yazar: {{ b.username }}</div>
                <div class="preview">
                  {{ (b.body || '').slice(0, 80) }}<span v-if="(b.body||'').length>80">…</span>
                </div>
              </div>

              <div class="item-actions">
                <button class="btn btn-green" @click="selectBlog(b)">Seç</button>
              </div>
            </li>
          </ul>

          <!-- Alt sayfalama (uzun listelerde faydalı) -->
          <div v-if="pendingCount > 0" class="pager bottom">
            <button class="btn btn-gray" @click="goFirst" :disabled="loading || page===1">⏮</button>
            <button class="btn btn-gray" @click="goPrev"  :disabled="loading || page===1">←</button>
            <span class="page-info">{{ page }} / {{ totalPages }}</span>
            <button class="btn btn-gray" @click="goNext"  :disabled="loading || page===totalPages">→</button>
            <button class="btn btn-gray" @click="goLast"  :disabled="loading || page===totalPages">⏭</button>
          </div>
        </div>

        <!-- Detay -->
        <div class="card">
          <div class="section-title">Detay</div>

          <div v-if="selected" class="detail">
            <div class="row">
              <div><b>Blog ID:</b> {{ selected.id ?? '—' }}</div>
              <div><b>Oluşturulma:</b> {{ selected.createdAt || '—' }}</div>
              <div><b>Güncellenme:</b> {{ selected.updatedAt || '—' }}</div>
            </div>

            <div class="row"><b>Başlık:</b> {{ selected.title }}</div>

            <div class="row flex">
              <div class="line">
                <div><b>Yazar:</b> {{ selected.username }}</div>
                <span class="badge badge-warn">pending</span>
              </div>
              <router-link :to="`/u/${encodeURIComponent(selected.username)}`">
                <button class="btn btn-gray">Yazar Profilini Aç</button>
              </router-link>
            </div>

            <div class="grid">
              <div><b>Durum:</b> {{ selected.status || '—' }}</div>
              <div><b>Onay:</b> {{ selected.isApproved ? 'Evet' : 'Hayır' }}</div>
              <div class="span-2"><b>Etiketler:</b> {{ selected.tags || '—' }}</div>
              <div class="span-2"><b>Kategori:</b> {{ selected.category || '—' }}</div>
            </div>

            <div class="row">
              <b>İçerik:</b>
              <div class="content">{{ selected.body }}</div>
            </div>

            <!-- Admin aksiyonları -->
            <div class="actions">
              <button
                  v-if="!selected.isApproved"
                  class="btn btn-green"
                  @click="approveSelected"
                  :disabled="actionLoading.approve"
              >
                {{ actionLoading.approve ? 'Onaylanıyor…' : 'Onayla' }}
              </button>

              <button
                  v-else
                  class="btn btn-gray"
                  @click="unapproveSelected"
                  :disabled="actionLoading.unapprove"
              >
                {{ actionLoading.unapprove ? 'Kaldırılıyor…' : 'Onayı Kaldır' }}
              </button>
            </div>
          </div>

          <p v-else class="muted">Listeden bir blog seçin.</p>
        </div>
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
.count { opacity: .7; font-weight: 600; }

.flow { display: grid; gap: 12px; }
.toolbar-card { padding: 12px 16px; }
.toolbar {
  display: flex; align-items: center; justify-content: space-between; gap: 12px;
}
.toolbar .left { display: flex; gap: 10px; align-items: center; }
.toolbar .right { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
.label {
  font-size: 13px; font-weight: 700; color: var(--color-heading); letter-spacing: .2px;
}

/* İki sütun düzeni */
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
.item:hover { border-color: var(--color-border-hover); background: var(--color-background-soft); }
.item-left { display: grid; gap: 6px; }
.line { display: flex; align-items: center; gap: 8px; }
.title { letter-spacing: .2px; }
.sub { opacity: .85; font-size: 13px; }
.preview { opacity: .9; margin-top: 2px; font-size: 13px; }
.item-actions { display: flex; gap: 8px; align-items: center; }

/* Alt/üst sayfalama */
.pager { display: inline-flex; gap: 6px; align-items: center; }
.pager.bottom { margin-top: 10px; }
.page-info { min-width: 70px; text-align: center; }

/* Detay */
.detail { display: grid; gap: 10px; }
.row { display: grid; gap: 4px; }
.flex { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.span-2 { grid-column: span 2; }
.content {
  margin-top: 6px;
  white-space: pre-wrap;
  border: 1px solid var(--color-border);
  border-radius: 10px;
  padding: 10px;
}

/* Butonlar (tema uyumlu) */
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
.badge-warn {
  color: #946200;
  border-color: rgba(148,98,0,.25);
  background: rgba(148,98,0,.08);
}

.muted { opacity: .8; }

/* Mobil */
@media (max-width: 900px) {
  .split { grid-template-columns: 1fr; }
  .toolbar { flex-direction: column; align-items: stretch; }
  .toolbar .right { justify-content: space-between; }
}
</style>
