<script setup>
import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import api from "../api/axios";

const router = useRouter();

const q = ref("");
const allBlogs = ref([]);
const selected = ref(null);

// admin kontrolü için
const me = ref({ role: "" });
const isAdmin = computed(() => me.value.role === "admin");

// --- butonlar için loading flag'leri
const loading = ref({
  approve: false,
  unapprove: false,
  restore: false,
});

/** MyAccount / BlogsAll ile aynı normalize */
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
        const { data } = await api.get(`/user/${encodeURIComponent(u)}`);
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

// q boşsa sonuç göstermeyelim
const results = computed(() => {
  const s = q.value.trim().toLowerCase();
  if (!s) return [];
  return allBlogs.value.filter(b => (b.title || "").toLowerCase().startsWith(s));
});

function selectBlog(b){ selected.value = b; }

// --- Admin aksiyonları (loading/disable ile) ---
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

// --- Yazar Profili: şimdilik /users sayfasına username query ile yönlendiriyoruz.
// UsersView bu query-param ile prefill yapmıyorsa yine de /users sayfasına gider.
// (İstersen UsersView’a query’den prefill etme özelliğini ekleriz.)
function goAuthorProfile() {
  if (!selected?.value?.username) return;
  router.push({ path: "/users", query: { username: selected.value.username } });
}
</script>

<template>
  <section style="max-width:1000px;margin:auto;display:grid;gap:16px">
    <h1>Blog Ara</h1>

    <div style="display:flex;gap:12px;align-items:center">
      <input
          v-model="q"
          placeholder="Başlıkla ara (en az 1 harf)…"
          style="flex:1;padding:8px;border:1px solid #ddd;border-radius:8px"
      />
      <router-link to="/blogs/all">
        <button style="padding:8px 12px;border:1px solid #333;border-radius:6px;cursor:pointer">
          Tüm Blogları Gör
        </button>
      </router-link>
    </div>

    <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px">
      <!-- Liste -->
      <div>
        <h3>Sonuçlar</h3>
        <p v-if="q && results.length === 0" style="opacity:.7">Eşleşen blog bulunamadı.</p>
        <ul v-else-if="q.trim()" style="display:grid;gap:6px">
          <li
              v-for="b in results"
              :key="b.id || b.title"
              style="display:flex;justify-content:space-between;align-items:center;border:1px solid #eee;padding:8px;border-radius:8px"
          >
            <div>
              <b>{{ b.title }}</b>
              <span v-if="b.status === 'deleted'" style="margin-left:8px;color:#b00;font-weight:bold">[silindi]</span>
              <span v-else style="margin-left:8px;font-size:12px;opacity:.8">[{{ b.isApproved ? 'approved' : 'pending' }}]</span>
              <div style="opacity:.7">Yazar: {{ b.username }}</div>
              <div style="opacity:.8; margin-top:4px">
                {{ (b.body || '').slice(0, 80) }}<span v-if="(b.body||'').length>80">…</span>
              </div>
            </div>
            <button @click="selectBlog(b)">Seç</button>
          </li>
        </ul>
        <p v-else style="opacity:.7">Arama yapmaya başlayın…</p>
      </div>

      <!-- Detay -->
      <div>
        <h3>Detay</h3>
        <div v-if="selected" style="display:grid;gap:8px;border:1px solid #eee;padding:12px;border-radius:8px">
          <div><b>Blog ID:</b> {{ selected.id ?? '—' }}</div>
          <div><b>Oluşturulma:</b> {{ selected.createdAt }}</div>
          <div><b>Güncellenme:</b> {{ selected.updatedAt }}</div>
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
          <div><b>Onay:</b> {{ selected.isApproved ? 'Evet' : 'Hayır' }}</div>
          <div><b>Durum:</b> {{ selected.status || '—' }}</div>
          <div><b>Kategori:</b> {{ selected.category || '—' }}</div>
          <div><b>Etiketler:</b> {{ selected.tags || '—' }}</div>
          <div><b>İçerik:</b></div>
          <div style="white-space:pre-wrap">{{ selected.body }}</div>

          <!-- Admin aksiyonları + loading -->
          <div v-if="isAdmin" style="margin-top:12px;display:flex;gap:8px;flex-wrap:wrap">
            <button
                v-if="selected.status === 'deleted'"
                @click="restoreSelected"
                :disabled="loading.restore"
            >
              {{ loading.restore ? 'Geri Yükleniyor…' : 'Geri Yükle' }}
            </button>

            <template v-else>
              <button
                  v-if="!selected.isApproved"
                  @click="approveSelected"
                  :disabled="loading.approve"
              >
                {{ loading.approve ? 'Onaylanıyor…' : 'Onayla' }}
              </button>

              <button
                  v-else
                  @click="unapproveSelected"
                  :disabled="loading.unapprove"
              >
                {{ loading.unapprove ? 'Kaldırılıyor…' : 'Onayı Kaldır' }}
              </button>
            </template>
          </div>
        </div>
        <p v-else>Arama yapın ve listeden bir blog seçin.</p>
      </div>
    </div>
  </section>
</template>
