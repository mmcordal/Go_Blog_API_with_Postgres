<!-- client/src/views/HomeView.vue -->
<script setup>
import { ref, onMounted, computed } from "vue";
import { RouterLink } from "vue-router";
import api from "../api/axios";

const role = ref(localStorage.getItem("role") || "");
const me = ref({ username: localStorage.getItem("username") || "" });
const recent = ref([]);

// Basit normalize (onaylı ve silinmemiş filtrele)
function normalize(b) {
  const delObj =
      b.deletedAt ?? b.DeletedAt ?? b.deleted_at ?? b.baseModel?.deletedAt ?? b.BaseModel?.DeletedAt ?? null;
  const deletedAt = (delObj && (delObj.Time || delObj.time || delObj)) || null;
  const isDeleted = !!delObj && (delObj.Valid === true || !!delObj.Time || !!delObj.time || !!deletedAt);

  const isApprovedRaw = b.is_approved ?? b.isApproved ?? b.content?.is_approved ?? b.content?.isApproved ?? false;

  return {
    id: b.id ?? b.ID ?? b.baseModel?.id ?? b.BaseModel?.ID ?? null,
    title: b.title ?? b.content?.title ?? "",
    body: b.body ?? b.content?.body ?? "",
    username: b.username ?? b.content?.username ?? "",
    isApproved: isDeleted ? false : !!isApprovedRaw,
    status: isDeleted ? "deleted" : (b.status ?? b.content?.status ?? ""),
    createdAt: b.createdAt ?? b.created_at ?? "",
  };
}

const greeting = computed(() =>
    me.value.username ? `Hoş geldin, @${me.value.username}` : "LogNode’a hoş geldin"
);

onMounted(async () => {
  try {
    // /blogs → backend’de onay/silinmiş filtreleri karışık olabilir, biz client’ta filtreliyoruz
    const { data } = await api.get("/blogs");
    const all = (data?.data || []).map(normalize);
    recent.value = all
        .filter(b => b.isApproved && b.status !== "deleted")
        .slice(0, 4);
  } catch (e) {
    // sessiz geç – home kritik değil
    console.warn("recent blogs failed:", e?.response?.data || e?.message);
    recent.value = [];
  }
});

function canCreate() {
  return role.value === "writer" || role.value === "admin";
}
function excerpt(t, n = 120) {
  if (!t) return "";
  return t.length > n ? t.slice(0, n) + "…" : t;
}
</script>

<template>
  <section class="hero">
    <div class="glow"></div>
    <div class="hero-inner">
      <img src="@/assets/logo-lognode.svg" alt="LogNode" class="hero-logo" />
      <h1 class="hero-title">LogNode</h1>
      <p class="hero-sub">
        {{ greeting }} — modern, sade ve güçlü bir blog platformu.
      </p>
      <div class="hero-actions">
        <RouterLink to="/blogs" class="btn">Blogları Keşfet</RouterLink>
        <RouterLink v-if="canCreate()" to="/blog-create" class="btn btn-primary">Yazmaya Başla</RouterLink>
      </div>
    </div>
  </section>

  <section class="recent container" v-if="recent.length">
    <div class="recent-head">
      <h2>En Son Onaylanmış Yazılar</h2>
      <RouterLink to="/blogs" class="see-all">Tümünü Gör</RouterLink>
    </div>

    <ul class="cards">
      <li v-for="b in recent" :key="b.id || b.title" class="card">
        <div class="card-top">
          <span class="badge">approved</span>
          <h3 class="card-title">{{ b.title }}</h3>
          <div class="meta">by <RouterLink :to="`/u/${encodeURIComponent(b.username)}`" class="meta-link">@{{ b.username }}</RouterLink></div>
        </div>
        <p class="card-excerpt">{{ excerpt(b.body) }}</p>
        <div class="card-actions">
          <RouterLink to="/blogs" class="btn-sm">Bloglarda Aç</RouterLink>
          <RouterLink :to="`/u/${encodeURIComponent(b.username)}`" class="btn-sm ghost">Yazar Profili</RouterLink>
        </div>
      </li>
    </ul>
  </section>
</template>

<style scoped>
/* ---- HERO ---- */
.hero {
  position: relative;
  overflow: hidden;
  padding: 56px 0 52px;
  background: linear-gradient(135deg, #0e0f12 0%, #0f1b16 100%);
  border: 1px solid var(--color-border);
  border-radius: 16px;
}

.hero-inner {
  position: relative;
  z-index: 2;
  max-width: 980px;
  margin: 0 auto;
  padding: 0 16px;
  display: grid;
  place-items: center;
  text-align: center;
  gap: 10px;
}

.glow {
  position: absolute;
  inset: -20% -30% auto -30%;
  height: 380px;
  background:
      radial-gradient(400px 200px at 50% 70%, rgba(25,210,124,0.18), transparent 60%),
      radial-gradient(500px 240px at 70% -10%, rgba(25,210,124,0.08), transparent 70%);
  filter: blur(20px);
}

.hero-logo { width: 72px; height: 72px; opacity: .95; }
.hero-title {
  font-size: 40px;
  font-weight: 800;
  letter-spacing: .3px;
  color: var(--color-heading);
}
.hero-sub {
  max-width: 680px;
  opacity: .9;
  line-height: 1.7;
}

.hero-actions {
  margin-top: 8px;
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  justify-content: center;
}
.btn,
.btn-primary {
  padding: 10px 14px;
  border-radius: 10px;
  border: 1px solid var(--color-border);
  text-decoration: none;
  color: var(--color-text);
  transition: all .2s ease;
}
.btn:hover { background: var(--color-background-soft); }
.btn-primary {
  border-color: hsla(160, 100%, 37%, 0.45);
  box-shadow: 0 0 0 2px hsla(160,100%,37%,0.18) inset;
}
.btn-primary:hover {
  background: rgba(25,210,124,0.08);
}

/* ---- RECENT ---- */
.container {
  max-width: 1100px;
  margin: 0 auto;
  padding: 22px 16px 0;
}

.recent { margin-top: 10px; }
.recent-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}
.recent h2 { font-size: 20px; }
.see-all {
  font-size: 14px;
  text-decoration: none;
  color: var(--color-text);
  border: 1px solid var(--color-border);
  padding: 6px 10px;
  border-radius: 8px;
  transition: all .2s ease;
}
.see-all:hover { background: var(--color-background-soft); }

.cards {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(1, minmax(0, 1fr));
}
@media (min-width: 720px) {
  .cards { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
@media (min-width: 1024px) {
  .cards { grid-template-columns: repeat(4, minmax(0, 1fr)); }
}

.card {
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 12px;
  background: rgba(12, 14, 13, 0.25);
  display: grid;
  gap: 8px;
}
.card-top { display: grid; gap: 6px; }
.badge {
  align-self: start;
  display: inline-block;
  padding: 2px 8px;
  font-size: 11px;
  border-radius: 999px;
  border: 1px solid rgba(25,210,124,0.45);
  color: #19d27c;
  background: rgba(25,210,124,0.08);
}
.card-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--color-heading);
}
.meta { font-size: 12px; opacity: .8; }
.meta-link {
  color: var(--color-text);
  text-decoration: none;
  border-bottom: 1px dashed rgba(25,210,124,0.35);
}
.meta-link:hover { border-bottom-style: solid; }

.card-excerpt {
  font-size: 14px;
  opacity: .9;
  min-height: 44px;
}

.card-actions {
  display: flex;
  gap: 8px;
  margin-top: 2px;
}
.btn-sm,
.btn-sm.ghost {
  padding: 6px 10px;
  border-radius: 8px;
  border: 1px solid var(--color-border);
  text-decoration: none;
  color: var(--color-text);
  font-size: 13px;
  transition: all .2s ease;
}
.btn-sm:hover { background: var(--color-background-soft); }
.btn-sm.ghost {
  background: transparent;
  border-color: transparent;
}
.btn-sm.ghost:hover { border-color: var(--color-border-hover); background: var(--color-background-soft); }
</style>
