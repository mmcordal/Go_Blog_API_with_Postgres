<script setup>
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import api from "../api/axios";

const router = useRouter();

// --- profil alanları ---
const me = ref(null);
const form = ref({ username: "", email: "", password: "" });

// --- Bloglar ---
const myBlogs = ref([]);
const selectedBlog = ref(null);

// düzenleme formu
const edit = ref({
  title: "",
  body: "",
  type: "",
  tags: "",
  category: "",
  status: "",
});

/** Tek noktadan sağlam normalize */
function normalizeBlog(b) {
  const deletedAtObj = b.deletedAt ?? b.DeletedAt ?? b.deleted_at ?? null;
  const deletedAt =
      (deletedAtObj && (deletedAtObj.Time || deletedAtObj.time || deletedAtObj)) || null;

  const isDeleted =
      !!deletedAtObj &&
      (deletedAtObj.Valid === true || !!deletedAtObj.Time || !!deletedAtObj.time || !!deletedAt);

  const authorId =
      b.author_id ?? b.authorId ?? b.content?.author_id ?? b.content?.authorId ?? null;

  const isApproved =
      typeof b.is_approved !== "undefined"
          ? b.is_approved
          : (b.content?.is_approved ?? false);

  return {
    id: b.id ?? b.ID ?? b.baseModel?.id ?? b.BaseModel?.ID ?? null,
    title: b.title ?? b.content?.title ?? "",
    body: b.body ?? b.content?.body ?? "",
    username: b.username ?? b.content?.username ?? "",
    type: b.type ?? b.content?.type ?? "",
    authorId,
    isApproved: isDeleted ? false : !!isApproved,
    status: isDeleted ? "deleted" : (b.status ?? b.content?.status ?? ""),
    tags: b.tags ?? "",
    category: b.category ?? "",
    createdAt: b.createdAt ?? b.created_at ?? "",
    updatedAt: b.updatedAt ?? b.updated_at ?? "",
    deletedAt: deletedAt || "",
    raw: b,
  };
}

// ---- LIFECYCLE
onMounted(async () => {
  const token = localStorage.getItem("token");
  if (!token) { router.push("/login"); return; }

  try {
    const [meRes, blogsRes] = await Promise.all([
      api.get("/me"),
      api.get("/blogs-deleted/me"),
    ]);

    me.value = meRes.data?.data || null;
    form.value.username = me.value?.username || "";
    form.value.email    = me.value?.email    || "";

    myBlogs.value = (blogsRes.data?.data || []).map(normalizeBlog);
  } catch (e) {
    console.error(e);
  }
});

// ---- METHODS
async function save() {
  if (!me.value) return;
  try {
    const body = {
      username: form.value.username,
      email:    form.value.email,
      ...(form.value.password ? { password: form.value.password } : {}),
    };
    await api.put("/me", body);
    alert("Profil güncellendi. Tekrar giriş yapmanız gerekmektedir.");
    localStorage.setItem("username", form.value.username);
    localStorage.setItem("email",    form.value.email);
    form.value.password = "";
    logout();
  } catch (e) {
    alert(e?.response?.data?.error || "Güncelleme başarısız");
  }
}

async function removeAccount() {
  if (!me.value) return;
  if (!confirm("Hesabınızı silmek istediğinize emin misiniz?")) return;
  try {
    await api.delete("/me");
    alert("Hesap silindi.");
    localStorage.clear();
    router.push("/login");
  } catch (e) {
    alert(e?.response?.data?.error || "Silme başarısız");
  }
}

function selectMine(b) {
  selectedBlog.value = b;
  edit.value = {
    title: b.title || "",
    body: b.body || "",
    type: b.type || "",
    tags: b.tags || "",
    category: b.category || "",
    status: b.status || "",
  };
}

function canEditSelectedBlog() {
  if (!selectedBlog.value || !me.value) return false;
  if (selectedBlog.value.status === "deleted") return false;
  return me.value.role === "admin" || selectedBlog.value.username === me.value.username;
}

async function updateSelectedBlog() {
  if (!selectedBlog.value) return;
  try {
    const oldTitle = selectedBlog.value.title;
    const payload = { ...edit.value };
    await api.put(`/blog/${encodeURIComponent(oldTitle)}`, payload);

    selectedBlog.value = { ...selectedBlog.value, ...payload, updatedAt: new Date().toISOString() };
    const i = myBlogs.value.findIndex(
        (x) => (x.id ?? x.title) === (selectedBlog.value.id ?? oldTitle)
    );
    if (i !== -1) myBlogs.value[i] = { ...myBlogs.value[i], ...payload };
    alert("Blog güncellendi.");
  } catch (e) {
    alert(e?.response?.data?.error || "Güncelleme başarısız");
  }
}

async function deleteSelectedBlog() {
  if (!selectedBlog.value) return;
  if (!confirm(`Silinsin mi? (${selectedBlog.value.title})`)) return;
  try {
    await api.delete(`/blog/${encodeURIComponent(selectedBlog.value.title)}`);

    const i = myBlogs.value.findIndex(
        (b) => (b.id ?? b.title) === (selectedBlog.value.id ?? selectedBlog.value.title)
    );
    if (i !== -1) {
      myBlogs.value[i] = {
        ...myBlogs.value[i],
        status: "deleted",
        isApproved: false,
        deletedAt: new Date().toISOString(),
      };
      selectedBlog.value = { ...myBlogs.value[i] };
    }
    alert("Blog silindi (soft delete).");
    selectedBlog.value = null;
  } catch (e) {
    alert(e?.response?.data?.error || "Silme başarısız");
  }
}

function logout() {
  localStorage.clear();
  window.dispatchEvent(new Event("auth:changed"));
  window.location.href = "/login";
}
</script>

<template>
  <section class="page">
    <div class="bg-glow" aria-hidden="true"></div>

    <header class="page-head">
      <div class="head-left">
        <img src="@/assets/logo-lognode.svg" class="brand" alt="LogNode" />
        <div>
          <h1>Hesabım</h1>
          <p class="muted">Profil bilgilerin ve tüm blogların tek ekranda.</p>
        </div>
      </div>

      <div v-if="me" class="stats">
        <div class="stat">
          <span class="label">Kullanıcı</span>
          <span class="value">@{{ me.username }}</span>
        </div>
        <div class="stat">
          <span class="label">Rol</span>
          <span class="chip chip-role">{{ me.role }}</span>
        </div>
        <div class="stat">
          <span class="label">Blog</span>
          <span class="value">{{ myBlogs.length }}</span>
        </div>
      </div>
    </header>

    <!-- Profil Kartı -->
    <div class="cards">
      <div class="card">
        <h3 class="card-title">Profil</h3>
        <div v-if="me" class="form">
          <div class="grid-2">
            <div>
              <label for="u">Kullanıcı adı</label>
              <input id="u" v-model="form.username" />
            </div>
            <div>
              <label for="e">Email</label>
              <input id="e" v-model="form.email" />
            </div>
          </div>

          <label for="p">Yeni şifre (opsiyonel)</label>
          <input id="p" v-model="form.password" type="password" placeholder="••••••••" />

          <div class="row">
            <button class="btn btn-green" @click="save">Kaydet</button>
            <button class="btn btn-red"   @click="removeAccount">Hesabı Sil</button>
          </div>
        </div>
        <p v-else>Yükleniyor...</p>
      </div>

      <!-- Bloglar -->
      <div class="card stretch">
        <h3 class="card-title">Bloglarım <span class="muted">(silinenler dahil)</span></h3>

        <div class="grid">
          <!-- Liste -->
          <div class="pane">
            <ul class="list">
              <li
                  v-for="b in myBlogs"
                  :key="b.id || b.title"
                  class="item"
              >
                <div class="item-main">
                  <div class="title-row">
                    <b class="title">{{ b.title }}</b>
                    <span
                        v-if="b.status === 'deleted'"
                        class="chip chip-danger"
                    >silindi</span>
                    <span
                        v-else
                        class="chip"
                        :class="b.isApproved ? 'chip-ok' : 'chip-warn'"
                    >{{ b.isApproved ? 'approved' : 'pending' }}</span>
                  </div>
                  <div class="meta">
                    <span class="by">Yazar: {{ b.username }}</span>
                    <span v-if="b.category" class="dot">•</span>
                    <span v-if="b.category" class="by">Kategori: {{ b.category }}</span>
                  </div>
                  <div class="snippet">
                    {{ (b.body || '').slice(0, 100) }}<span v-if="(b.body || '').length > 100">…</span>
                  </div>
                </div>

                <button
                    class="btn btn-green"
                    style="border:1px solid #333;border-radius:6px;padding:6px 10px;cursor:pointer"
                    @click="selectMine(b)"
                >
                  Seç
                </button>
              </li>
            </ul>
          </div>

          <!-- Detay + düzenleme -->
          <div class="pane">
            <h4 class="pane-title">Detay</h4>

            <div v-if="selectedBlog" class="detail">
              <div class="kv"><b>Blog ID:</b> <span>{{ selectedBlog.id ?? "—" }}</span></div>
              <div class="kv"><b>Oluşturulma:</b> <span>{{ selectedBlog.createdAt || "—" }}</span></div>
              <div class="kv"><b>Güncellenme:</b> <span>{{ selectedBlog.updatedAt || "—" }}</span></div>
              <div class="kv" v-if="selectedBlog.deletedAt"><b>Silinme (soft):</b> <span>{{ selectedBlog.deletedAt }}</span></div>

              <div class="kv"><b>Başlık:</b> <span>{{ selectedBlog.title }}</span></div>
              <div class="kv"><b>Yazar Username:</b> <span>{{ selectedBlog.username }}</span></div>
              <div class="kv"><b>AuthorID:</b> <span>{{ selectedBlog.authorId ?? "—" }}</span></div>
              <div class="kv"><b>Tip:</b> <span>{{ selectedBlog.type || "—" }}</span></div>
              <div class="kv"><b>Durum:</b>
                <span v-if="selectedBlog.status === 'deleted'" class="chip chip-danger">deleted</span>
                <span v-else class="chip" :class="selectedBlog.isApproved ? 'chip-ok' : 'chip-warn'">
                  {{ selectedBlog.isApproved ? "approved" : "pending" }}
                </span>
              </div>
              <div class="kv"><b>Etiketler:</b> <span>{{ selectedBlog.tags || "—" }}</span></div>
              <div class="kv"><b>Kategori:</b> <span>{{ selectedBlog.category || "—" }}</span></div>

              <div class="kv col">
                <b>İçerik:</b>
                <div class="content">{{ selectedBlog.body }}</div>
              </div>

              <!-- Güncelle / Sil sadece yazar veya admin & silinmemiş -->
              <div
                  v-if="canEditSelectedBlog()"
                  class="edit"
              >
                <h4>Güncelle</h4>

                <label>Başlık</label>
                <input v-model="edit.title" />

                <label>İçerik</label>
                <textarea v-model="edit.body" rows="6"></textarea>

                <div class="grid-2">
                  <div>
                    <label>Tip</label>
                    <input v-model="edit.type" />
                  </div>
                  <div>
                    <label>Durum</label>
                    <input v-model="edit.status" />
                  </div>
                </div>

                <div class="grid-2">
                  <div>
                    <label>Etiketler</label>
                    <input v-model="edit.tags" />
                  </div>
                  <div>
                    <label>Kategori</label>
                    <input v-model="edit.category" />
                  </div>
                </div>

                <div class="row">
                  <button
                      class="btn btn-green"
                      style="border:1px solid #333;border-radius:6px;padding:6px 10px;cursor:pointer"
                      @click="updateSelectedBlog"
                      :disabled="selectedBlog.status==='deleted'"
                  >Kaydet
                  </button>
                  <button
                      class="btn btn-red"
                      style="border:1px solid #333;border-radius:6px;padding:6px 10px;cursor:pointer"
                      @click="deleteSelectedBlog"
                      :disabled="selectedBlog.status==='deleted'"
                  >
                    Sil
                  </button>
                </div>
              </div>
            </div>

            <p v-else class="muted">Listeden bir blog seçin.</p>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
/* Sayfa zemin + başlık */
.page {
  position: relative;
  padding: 20px 0 32px;
}
.bg-glow {
  position: absolute;
  inset: 0;
  background:
      radial-gradient(1000px 600px at -10% -10%, rgba(25,210,124,0.12), transparent 60%),
      radial-gradient(700px 400px at 110% 10%, rgba(25,210,124,0.08), transparent 60%);
  z-index: -1;
}
.page-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  margin-bottom: 18px;
  border-bottom: 1px solid var(--color-border);
  padding-bottom: 12px;
}
.head-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.brand { width: 34px; height: 34px; }
h1 { font-size: 22px; line-height: 1.2; color: var(--color-heading); }
.muted { opacity: .8; font-size: 13px; }

/* Kısa istatistikler */
.stats { display: flex; gap: 14px; flex-wrap: wrap; }
.stat { display: grid; gap: 4px; text-align: right; }
.label { font-size: 11px; opacity: .7; }
.value { font-weight: 700; }

/* Chip’ler */
.chip {
  display: inline-block;
  padding: 2px 8px;
  font-size: 11px;
  border-radius: 999px;
  border: 1px solid var(--color-border);
  background: var(--color-background-soft);
}
.chip-ok { border-color: rgba(25,210,124,.55); background: rgba(25,210,124,.12); }
.chip-warn { border-color: rgba(255,200,80,.4); background: rgba(255,200,80,.12); }
.chip-danger { border-color: rgba(255,80,80,.45); background: rgba(255,80,80,.12); color: #ff7b7b; }
.chip-role { border-color: rgba(25,210,124,.5); background: rgba(25,210,124,.14); }

/* Kart alanı */
.cards {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
}
@media (min-width: 980px) {
  .cards {
    grid-template-columns: 1fr;
    gap: 18px;
  }
}
.card {
  background: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: 14px;
  box-shadow: 0 10px 40px rgba(0,0,0,0.28);
  padding: 16px;
}
.card.stretch { padding-bottom: 6px; }
.card-title {
  margin-bottom: 10px;
  font-size: 16px;
  font-weight: 700;
  color: var(--color-heading);
}

/* Form */
.form { display: grid; gap: 10px; }
.grid-2 { display: grid; grid-template-columns: 1fr; gap: 10px; }
@media (min-width: 720px) {
  .grid-2 { grid-template-columns: 1fr 1fr; }
}
label { font-size: 12px; font-weight: 600; margin-bottom: 4px; color: var(--color-heading); }
input, textarea {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--color-border);
  background: var(--color-background-soft);
  color: var(--color-text);
  border-radius: 10px;
  outline: none;
  transition: border-color .18s ease, box-shadow .18s ease, background .18s ease;
}
textarea { resize: vertical; }
input:focus, textarea:focus {
  border-color: var(--color-border-hover);
  box-shadow: 0 0 0 2px rgba(25,210,124,0.14) inset;
  background: var(--color-background);
}
.row { display: flex; gap: 10px; margin-top: 6px; flex-wrap: wrap; }

/* Blog grid */
.grid {
  display: grid;
  gap: 14px;
}
@media (min-width: 980px) {
  .grid {
    grid-template-columns: 1fr 1fr;
  }
}
.pane { border: 1px dashed var(--color-border); border-radius: 12px; padding: 10px; }
.pane-title { font-size: 14px; font-weight: 700; margin-bottom: 8px; }

/* Liste */
.list { display: grid; gap: 8px; }
.item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 10px;
  background: var(--color-background);
}
.item-main { flex: 1; min-width: 0; }
.title-row { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
.title { font-weight: 700; }
.meta { font-size: 12px; opacity: .8; display: flex; align-items: center; gap: 6px; }
.dot { opacity: .5; }
.snippet { opacity: .92; margin-top: 4px; }

/* Detay */
.detail { display: grid; gap: 8px; }
.kv { display: flex; gap: 8px; align-items: baseline; }
.kv.col { display: grid; gap: 6px; }
.content {
  white-space: pre-wrap;
  border: 1px solid var(--color-border);
  background: var(--color-background-soft);
  border-radius: 10px;
  padding: 10px;
  max-height: 260px;
  overflow: auto;
}

/* Küçük pürüzler */
button[disabled] { opacity: .6; cursor: default !important; }

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
</style>
