<script setup>
import { ref, onMounted, computed } from "vue";
import api from "../api/axios";

const me = ref({ role: "" });
const isAdmin = computed(() => me.value.role === "admin");

const all = ref([]);
const pending = computed(() =>
    all.value.filter(b => !b.isApproved && b.status !== "deleted")
);

const selected = ref(null);

// normalize her yerde kullandığımızla aynı mantık
function normalize(b){
  const deletedAtObj =
      b.deletedAt ?? b.DeletedAt ?? b.deleted_at ?? b.baseModel?.deletedAt ?? b.BaseModel?.DeletedAt ?? null;
  const deletedAt =
      (deletedAtObj && (deletedAtObj.Time || deletedAtObj.time || deletedAtObj)) || null;
  const isDeleted =
      !!deletedAtObj && (deletedAtObj.Valid === true || !!deletedAtObj.Time || !!deletedAtObj.time || !!deletedAt);

  const isApprovedRaw =
      b.is_approved ?? b.isApproved ?? b.content?.is_approved ?? b.content?.isApproved ?? false;
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
  };
}

onMounted(load);

async function load(){
  // admin mi?
  const u = localStorage.getItem("username");
  if (u) {
    try {
      const { data } = await api.get(`/user/${encodeURIComponent(u)}`);
      me.value.role = data?.data?.role || "";
    } catch {}
  }
  if (!isAdmin.value) return;

  // admin → include_deleted=true ile çek, sonra client’tan filtrele
  const { data } = await api.get("/blogs", { params: { include_deleted: true } });
  all.value = (data?.data || []).map(normalize);
}

function selectBlog(b){ selected.value = b; }

// --- admin aksiyonları ---
async function approveSelected() {
  if (!selected.value) return;
  await api.put(`/blog/${encodeURIComponent(selected.value.title)}/approve`);
  selected.value.isApproved = true;
  // listeden de çıkar (artık pending değil)
  all.value = all.value.map(x =>
      (x.id ?? x.title) === (selected.value.id ?? selected.value.title)
          ? { ...x, isApproved: true }
          : x
  );
}

async function unapproveSelected() {
  if (!selected.value) return;
  await api.put(`/blog/${encodeURIComponent(selected.value.title)}/unapprove`);
  selected.value.isApproved = false;
  // pending listesinde kalsın
  all.value = all.value.map(x =>
      (x.id ?? x.title) === (selected.value.id ?? selected.value.title)
          ? { ...x, isApproved: false }
          : x
  );
}
</script>

<template>
  <section style="max-width:1000px;margin:auto;display:grid;gap:16px">
    <h1>Onay Bekleyen Bloglar</h1>

    <div v-if="!isAdmin" style="opacity:.8">Bu sayfayı sadece adminler görebilir.</div>

    <div v-else style="display:grid;grid-template-columns:1fr 1fr;gap:16px">
      <!-- Liste -->
      <div>
        <h3>Bekleyenler ({{ pending.length }})</h3>
        <ul style="display:grid;gap:6px">
          <li
              v-for="b in pending"
              :key="b.id || b.title"
              style="display:flex;justify-content:space-between;align-items:center;border:1px solid #eee;padding:8px;border-radius:8px"
          >
            <div>
              <b>{{ b.title }}</b>
              <span style="margin-left:8px;font-size:12px;opacity:.8">[pending]</span>
              <div style="opacity:.7">Yazar: {{ b.username }}</div>
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
        <div v-if="selected" style="display:grid;gap:8px;border:1px solid #eee;padding:12px;border-radius:8px">
          <div><b>Blog ID:</b> {{ selected.id ?? '—' }}</div>
          <div><b>Başlık:</b> {{ selected.title }}</div>
          <div><b>Yazar:</b> {{ selected.username }}</div>
          <div><b>Durum:</b> {{ selected.status || '—' }}</div>
          <div><b>Onay:</b> {{ selected.isApproved ? 'Evet' : 'Hayır' }}</div>
          <div><b>Oluşturulma:</b> {{ selected.createdAt }}</div>
          <div><b>Güncellenme:</b> {{ selected.updatedAt }}</div>
          <div><b>İçerik:</b></div>
          <div style="white-space:pre-wrap">{{ selected.body }}</div>

          <div style="margin-top:12px;display:flex;gap:8px;flex-wrap:wrap">
            <button v-if="!selected.isApproved" @click="approveSelected">Onayla</button>
            <button v-else @click="unapproveSelected">Onayı Kaldır</button>
          </div>
        </div>
        <p v-else>Listeden bir blog seçin.</p>
      </div>
    </div>
  </section>
</template>
