<script setup>
import { ref, onMounted, computed } from "vue";
import api from "../api/axios";

/* --- State --- */
const me = ref({ role: "" });
const isAdmin = computed(() => (me.value.role || "").toLowerCase() === "admin");

const status = ref("pending"); // pending | approved | rejected
const all = ref([]);
const selected = ref(null);

const loading = ref(false);
const actionLoading = ref({ approve: false, reject: false });
const fetchError = ref("");

/* --- Helpers --- */
function normalize(r = {}) {
  return {
    id: r.id ?? r.ID ?? r.request_id ?? r.RequestID ?? "",
    username: r.username ?? r.user_name ?? "",
    requested_role: r.requested_role ?? r.requestedRole ?? "",
    status: r.status ?? "",
    reason: r.reason || "",
    decided_by: r.decided_by || r.decidedBy || "",
    decided_at: r.decided_at || r.decidedAt || "",
    created_at: r.created_at || r.createdAt || "",
  };
}

/* --- Load --- */
async function load() {
  fetchError.value = "";
  loading.value = true;
  selected.value = null;

  try {
    // rol kontrolü
    try {
      const { data } = await api.get("/me");
      me.value.role = data?.data?.role || "";
    } catch {
      me.value.role = "";
    }

    if (!isAdmin.value) {
      all.value = [];
      return; // admin değilse liste çağrısı yapma
    }

    const { data } = await api.get("/role-requests", {
      params: { status: status.value },
    });
    all.value = (data?.data || []).map(normalize);
  } catch (e) {
    fetchError.value = e?.response?.data?.error || e?.message || "Liste alınamadı.";
    all.value = [];
  } finally {
    loading.value = false;
  }
}

onMounted(load);

function onStatusChange() {
  selected.value = null;
  load();
}

/* --- Actions --- */
async function approveSelected() {
  if (!selected.value || actionLoading.value.approve) return;
  actionLoading.value.approve = true;
  try {
    await api.put(`/role-requests/${selected.value.id}/approve`);
    alert("Talep onaylandı.");
    await load();
  } catch (e) {
    alert(e?.response?.data?.error || "Onay başarısız");
  } finally {
    actionLoading.value.approve = false;
  }
}

async function rejectSelected() {
  if (!selected.value || actionLoading.value.reject) return;
  actionLoading.value.reject = true;
  try {
    await api.put(`/role-requests/${selected.value.id}/reject`);
    alert("Talep reddedildi.");
    await load();
  } catch (e) {
    alert(e?.response?.data?.error || "Ret başarısız");
  } finally {
    actionLoading.value.reject = false;
  }
}
</script>

<template>
  <section class="page">
    <div class="header">
      <h1>Rol Talepleri</h1>
      <p>Adminler, kullanıcıların rol yükseltme taleplerini buradan yönetebilir.</p>
    </div>

    <div v-if="!isAdmin" class="card muted">
      Bu sayfayı sadece adminler görebilir.
    </div>

    <div v-else class="flow">
      <!-- Toolbar -->
      <div class="card">
        <div class="toolbar">
          <div class="toolbar-left">
            <label class="label">Durum</label>
            <select v-model="status" @change="onStatusChange" class="select" :disabled="loading">
              <option value="pending">Bekleyen</option>
              <option value="approved">Onaylanan</option>
              <option value="rejected">Reddedilen</option>
            </select>
          </div>
          <button class="btn btn-gray" @click="load" :disabled="loading">
            {{ loading ? 'Yükleniyor…' : 'Yenile' }}
          </button>
        </div>
      </div>

      <div class="card" v-if="fetchError">
        <p class="muted">Hata: {{ fetchError }}</p>
      </div>

      <div class="split">
        <!-- Liste -->
        <div class="card">
          <div class="section-title">
            Liste <span class="muted">({{ all.length }})</span>
          </div>

          <p v-if="loading" class="muted">Yükleniyor…</p>

          <p v-else-if="all.length === 0" class="muted">
            Bu durumda talep yok.
          </p>

          <ul v-else class="list">
            <li v-for="r in all" :key="r.id" class="item">
              <div class="item-left">
                <div class="line">
                  <b class="title">@{{ r.username }}</b>
                  <span class="arrow">→</span>
                  <span
                      class="badge"
                      :class="{
                      'badge-success': r.status==='approved',
                      'badge-warn': r.status==='pending',
                      'badge-danger': r.status==='rejected'
                    }"
                  >
                    {{ r.requested_role }}
                  </span>
                </div>
                <div class="sub">
                  Durum:
                  <span
                      class="badge"
                      :class="{
                      'badge-success': r.status==='approved',
                      'badge-warn': r.status==='pending',
                      'badge-danger': r.status==='rejected'
                    }"
                  >
                    {{ r.status }}
                  </span>
                  <span v-if="r.reason" class="note">• Not: {{ r.reason }}</span>
                </div>
                <div class="sub light">Talep tarihi: {{ r.created_at }}</div>
              </div>

              <div class="item-actions">
                <button class="btn btn-green" @click="selected = r">Seç</button>
                <router-link
                    :to="`/u/${encodeURIComponent(r.username)}`"
                    :title="`${r.username} profili`"
                >
                  <button class="btn btn-gray">Profili Aç</button>
                </router-link>
              </div>
            </li>
          </ul>
        </div>

        <!-- Detay -->
        <div class="card">
          <div class="section-title">Detay</div>

          <div v-if="selected" class="detail">
            <div class="grid">
              <div><b>ID:</b> {{ selected.id }}</div>
              <div><b>Kullanıcı:</b> @{{ selected.username }}</div>
              <div><b>İstenen Rol:</b> {{ selected.requested_role }}</div>
              <div>
                <b>Durum:</b>
                <span
                    class="badge"
                    :class="{
                    'badge-success': selected.status==='approved',
                    'badge-warn': selected.status==='pending',
                    'badge-danger': selected.status==='rejected'
                  }"
                >
                  {{ selected.status }}
                </span>
              </div>
              <div class="span-2" v-if="selected.reason"><b>Not:</b> {{ selected.reason }}</div>
              <div v-if="selected.decided_by"><b>Karar veren:</b> {{ selected.decided_by }}</div>
              <div v-if="selected.decided_at"><b>Karar zamanı:</b> {{ selected.decided_at }}</div>
              <div><b>Talep Tarihi:</b> {{ selected.created_at }}</div>
            </div>

            <div v-if="status === 'pending'" class="actions">
              <button class="btn btn-green" @click="approveSelected" :disabled="actionLoading.approve">
                {{ actionLoading.approve ? 'Onaylanıyor…' : 'Onayla' }}
              </button>
              <button class="btn btn-red" @click="rejectSelected" :disabled="actionLoading.reject">
                {{ actionLoading.reject ? 'Reddediliyor…' : 'Reddet' }}
              </button>
            </div>
          </div>

          <p v-else class="muted">Listeden bir talep seçin.</p>
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
.flow { display: grid; gap: 12px; }

/* Toolbar */
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.toolbar-left { display: flex; align-items: center; gap: 10px; }
.label {
  font-size: 13px;
  font-weight: 700;
  color: var(--color-heading);
  letter-spacing: .2px;
}
.select {
  padding: 10px 12px;
  border: 1px solid var(--color-border);
  border-radius: 10px;
  background: var(--color-background);
  color: var(--color-text);
  outline: none;
  transition: border-color .15s ease, box-shadow .15s ease;
}
.select:focus {
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
.sub { opacity: .85; font-size: 13px; }
.sub.light { opacity: .7; font-size: 12px; }
.note { opacity: .85; }
.arrow { opacity: .6; }

/* Detay */
.detail { display: grid; gap: 12px; }
.grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
.span-2 { grid-column: span 2; }

/* Butonlar */
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
.btn-red {
  background: rgba(255,0,0,.06);
  border-color: rgba(255,0,0,.35);
  color: #ff6e6e;
}
.btn-red:hover { background: rgba(255,0,0,.1); }

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
  .toolbar { flex-direction: column; align-items: stretch; }
  .toolbar-left { justify-content: space-between; }
}
</style>
