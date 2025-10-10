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
  // GORM DeletedAt genelde {Time:"...", Valid:true} gibi gelir
  const deletedAtObj =
      b.deletedAt ??
      b.DeletedAt ??
      b.deleted_at ??
      null;

  const deletedAt =
      (deletedAtObj && (deletedAtObj.Time || deletedAtObj.time || deletedAtObj)) || null;

  const isDeleted =
      !!deletedAtObj &&
      (deletedAtObj.Valid === true || !!deletedAtObj.Time || !!deletedAtObj.time || !!deletedAt);

  const authorId =
      b.author_id ??
      b.authorId ??
      b.content?.author_id ??
      b.content?.authorId ??
      null;

  const isApproved =
      typeof b.is_approved !== "undefined"
          ? b.is_approved
          : (b.content?.is_approved ?? false);

  return {
    // kök ID (BaseModel.ID)
    id: b.id ?? b.ID ?? b.baseModel?.id ?? b.BaseModel?.ID ?? null,

    // temel alanlar
    title: b.title ?? b.content?.title ?? "",
    body: b.body ?? b.content?.body ?? "",
    username: b.username ?? b.content?.username ?? "",
    type: b.type ?? b.content?.type ?? "",

    authorId,
    isApproved: isDeleted ? false : !!isApproved, // silinmişse onaylı sayma
    status: isDeleted ? "deleted" : (b.status ?? b.content?.status ?? ""),
    tags: b.tags ?? "",
    category: b.category ?? "",

    // tarihler
    createdAt: b.createdAt ?? b.created_at ?? "",
    updatedAt: b.updatedAt ?? b.updated_at ?? "",
    deletedAt: deletedAt || "",

    raw: b,
  };
}

// sayfa yüklenince
onMounted(async () => {
  const myUsername = localStorage.getItem("username");
  if (!myUsername) {
    router.push("/login");
    return;
  }

  try {
    // profil
    const { data } = await api.get(`/user/${encodeURIComponent(myUsername)}`);
    me.value = data?.data || null;
    form.value.username = me.value.username;
    form.value.email = me.value.email;

    // ⚠️ aktif + silinmiş (soft delete) bloglar
    const res = await api.get(`/blogs-deleted/${encodeURIComponent(myUsername)}`);
    myBlogs.value = (res.data?.data || []).map(normalizeBlog);
  } catch (e) {
    console.error(e);
  }
});

// profil güncelle
async function save() {
  try {
    const body = { username: form.value.username, email: form.value.email };
    if (form.value.password) body.password = form.value.password;
    await api.put(`/user/${encodeURIComponent(me.value.username)}`, body);
    alert("Profil güncellendi. Tekrar giriş yapmanız gerekebilir.");
    localStorage.setItem("username", form.value.username);
    localStorage.setItem("email", form.value.email);
    form.value.password = "";
  } catch (e) {
    alert(e?.response?.data?.error || "Güncelleme başarısız");
  }
}

// hesap sil
async function removeAccount() {
  if (!confirm("Hesabınızı silmek istediğinize emin misiniz?")) return;
  try {
    await api.delete(`/user/${encodeURIComponent(me.value.username)}`);
    alert("Hesap silindi.");
    localStorage.clear();
    router.push("/login");
  } catch (e) {
    alert(e?.response?.data?.error || "Silme başarısız");
  }
}

// blog seç
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

// kim düzenleyebilir?
function canEditSelectedBlog() {
  if (!selectedBlog.value || !me.value) return false;
  if (selectedBlog.value.status === "deleted") return false; // silinmiş düzenlenemez
  return me.value.role === "admin" || selectedBlog.value.username === me.value.username;
}

// blog güncelle
async function updateSelectedBlog() {
  if (!selectedBlog.value) return;
  try {
    const oldTitle = selectedBlog.value.title;
    const payload = { ...edit.value };
    await api.put(`/blog/${encodeURIComponent(oldTitle)}`, payload);

    // UI’ı güncelle
    selectedBlog.value = {
      ...selectedBlog.value,
      ...payload,
      updatedAt: new Date().toISOString(),
    };
    const i = myBlogs.value.findIndex(
        (x) => (x.id ?? x.title) === (selectedBlog.value.id ?? oldTitle)
    );
    if (i !== -1) myBlogs.value[i] = { ...myBlogs.value[i], ...payload };
    alert("Blog güncellendi.");
  } catch (e) {
    alert(e?.response?.data?.error || "Güncelleme başarısız");
  }
}

// blog sil (soft delete)
async function deleteSelectedBlog() {
  if (!selectedBlog.value) return;
  if (!confirm(`Silinsin mi? (${selectedBlog.value.title})`)) return;
  try {
    await api.delete(`/blog/${encodeURIComponent(selectedBlog.value.title)}`);

    // Listeden tamamen kaldırmak yerine, "silindi" olarak işaretleyelim ki görünmeye devam etsin:
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
    }
    selectedBlog.value = { ...myBlogs.value[i] };
    alert("Blog silindi (soft delete).");
  } catch (e) {
    alert(e?.response?.data?.error || "Silme başarısız");
  }
}
</script>

<template>
  <section style="max-width:1000px;margin:auto;display:grid;gap:16px">
    <h1>Hesabım</h1>

    <!-- Profil -->
    <div v-if="me" style="display:grid;gap:8px;border:1px solid #eee;padding:12px;border-radius:8px">
      <div><b>ID:</b> {{ me.id }}</div>
      <div><b>Rol:</b> {{ me.role }}</div>

      <label>Kullanıcı adı</label>
      <input v-model="form.username" />

      <label>Email</label>
      <input v-model="form.email" />

      <label>Yeni şifre (opsiyonel)</label>
      <input v-model="form.password" type="password" />

      <div style="display:flex;gap:8px;margin-top:8px">
        <button @click="save">Kaydet</button>
        <button style="color:#b00" @click="removeAccount">Hesabı Sil</button>
      </div>
    </div>
    <p v-else>Yükleniyor...</p>

    <!-- Bloglar -->
    <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px">
      <!-- Liste -->
      <div>
        <h2>Bloglarım (silinenler dahil)</h2>
        <ul style="display:grid;gap:6px">
          <li
              v-for="b in myBlogs"
              :key="b.id || b.title"
              :style="{
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
              border: '1px solid #eee',
              padding: '8px',
              borderRadius: '8px',
              background: undefined
            }"
          >
            <div>
              <b>{{ b.title }}</b>
              <span
                  v-if="b.status === 'deleted'"
                  style="margin-left:8px;color:#b00;font-weight:bold"
              >
                [silindi]
              </span>
              <span v-else style="margin-left:8px;font-size:12px;opacity:.8">
                [{{ b.isApproved ? 'approved' : 'pending' }}]
              </span>
              <div style="opacity:.8;margin-top:4px">
                {{ (b.body || '').slice(0, 80) }}<span v-if="(b.body || '').length > 80">…</span>
              </div>
            </div>
            <button @click="selectMine(b)">Seç</button>
          </li>
        </ul>
      </div>

      <!-- Detay + düzenleme -->
      <div>
        <h3>Detay</h3>
        <div v-if="selectedBlog" style="display:grid;gap:8px;border:1px solid #eee;padding:12px;border-radius:8px">
          <div><b>Blog ID:</b> {{ selectedBlog.id ?? "—" }}</div>
          <div><b>Oluşturulma:</b> {{ selectedBlog.createdAt || "—" }}</div>
          <div><b>Güncellenme:</b> {{ selectedBlog.updatedAt || "—" }}</div>
          <div v-if="selectedBlog.deletedAt"><b>Silinme (soft):</b> {{ selectedBlog.deletedAt }}</div>

          <div><b>Başlık:</b> {{ selectedBlog.title }}</div>
          <div><b>Yazar Username:</b> {{ selectedBlog.username }}</div>
          <div><b>AuthorID:</b> {{ selectedBlog.authorId ?? "—" }}</div>
          <div><b>Tip:</b> {{ selectedBlog.type || "—" }}</div>
          <div><b>Durum:</b> {{ selectedBlog.status || "—" }}</div>
          <div><b>Onay:</b> {{ selectedBlog.isApproved ? "Evet" : "Hayır" }}</div>
          <div><b>Etiketler:</b> {{ selectedBlog.tags || "—" }}</div>
          <div><b>Kategori:</b> {{ selectedBlog.category || "—" }}</div>
          <div><b>İçerik:</b></div>
          <div style="white-space:pre-wrap">{{ selectedBlog.body }}</div>

          <!-- Güncelle / Sil sadece yazar veya admin & silinmemiş -->
          <div
              v-if="canEditSelectedBlog()"
              style="margin-top:12px;display:grid;gap:8px;border-top:1px solid #eee;padding-top:12px;"
          >
            <h4>Güncelle</h4>

            <label>Başlık</label>
            <input v-model="edit.title" />

            <label>İçerik</label>
            <textarea v-model="edit.body" rows="6"></textarea>

            <label>Tip</label>
            <input v-model="edit.type" />

            <label>Etiketler</label>
            <input v-model="edit.tags" />

            <label>Kategori</label>
            <input v-model="edit.category" />

            <label>Durum</label>
            <input v-model="edit.status" />

            <div style="display:flex;gap:8px;margin-top:8px">
              <button @click="updateSelectedBlog" :disabled="selectedBlog.status==='deleted'">
                Kaydet
              </button>
              <button
                  style="color:#b00"
                  @click="deleteSelectedBlog"
                  :disabled="selectedBlog.status==='deleted'"
              >
                Sil
              </button>
            </div>
          </div>
        </div>

        <p v-else>Listeden bir blog seçin.</p>
      </div>
    </div>
  </section>
</template>
