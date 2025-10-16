<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick } from "vue";
import { useRouter } from "vue-router";
import api from "../api/axios";

const router = useRouter();

// --- form meta ---
const TYPES = ["tech", "lifestyle", "note", "howto"];
const CATEGORIES = ["golang", "frontend", "backend", "database", "devops"];
const STATUSES = ["draft", "published"];

// --- rol kontrolü ---
const role = ref("");
onMounted(async () => {
  try {
    const { data } = await api.get("/me");
    role.value = data?.data?.role || "";
    if (role.value !== "writer" && role.value !== "admin") {
      alert("Blog oluşturmak için writer veya admin olmalısınız.");
      router.push("/blogs");
    }
  } catch {
    router.push("/login");
  }
});

// --- klasik alanlar ---
const form = ref({
  title: "",
  type: TYPES[0],
  tags: "",
  category: CATEGORIES[0],
  status: STATUSES[0],
});

// --- RICH EDITOR (görsel, DB'ye plain text gider) ---
const editorRef = ref(null);
const fileInputRef = ref(null);
const editorPlaceholder =
    "İçeriği buraya yazın. Üstteki araç çubuğu ile kalın/italik/altı çizili, hizalama, liste, görsel vb. uygulayın. Kaydederken düz metin gönderilir.";

const active = ref({
  bold: false,
  italic: false,
  underline: false,
  strikeThrough: false,
  justifyLeft: false,
  justifyCenter: false,
  justifyRight: false,
  justifyFull: false,
  unorderedList: false,
  orderedList: false,
});
const fontSizeState = ref(3); // 1..7
const foreColor = ref("#ffffff");

let selListener = null;

function focusEditor() {
  editorRef.value?.focus();
}

function apply(cmd, value = null) {
  focusEditor();
  document.execCommand(cmd, false, value);
  refreshState();
}

function toggleList(kind) {
  apply(kind === "ul" ? "insertUnorderedList" : "insertOrderedList");
}

function setAlign(where) {
  const map = {
    left: "justifyLeft",
    center: "justifyCenter",
    right: "justifyRight",
    full: "justifyFull",
  };
  apply(map[where]);
}

function setFontSize(delta) {
  const cur = Number(queryFontSize());
  let next = Math.min(7, Math.max(1, cur + delta));
  fontSizeState.value = next;
  apply("fontSize", String(next));
}

function setColor(color) {
  foreColor.value = color;
  apply("foreColor", color);
}

function insertSeparator() {
  apply("insertText", "\n\n---\n\n");
}

function insertLink() {
  focusEditor();
  const url = prompt("Bağlantı URL’si:");
  if (!url) return;
  document.execCommand("createLink", false, url);
  refreshState();
}

// --- Görsel ekleme (URL) ---
function insertImageByURL() {
  const url = prompt("Görsel URL’si:");
  if (!url) return;
  focusEditor();
  document.execCommand("insertImage", false, url);
  normalizeImages();
}

// --- Görsel ekleme (Dosyadan) ---
function triggerFilePick() {
  fileInputRef.value?.click();
}

function onPickFile(e) {
  const file = e.target.files?.[0];
  if (!file) return;
  const reader = new FileReader();
  reader.onload = () => {
    focusEditor();
    document.execCommand("insertImage", false, reader.result); // base64 dataURL
    normalizeImages();
    e.target.value = ""; // aynı dosyayı tekrar seçebilsin
  };
  reader.readAsDataURL(file);
}

// --- Paste ile görsel yapıştırma desteği (clipboard image) ---
function onPaste(e) {
  const items = e.clipboardData?.items || [];
  for (const it of items) {
    if (it.type && it.type.startsWith("image/")) {
      const file = it.getAsFile();
      if (!file) continue;
      e.preventDefault();
      const reader = new FileReader();
      reader.onload = () => {
        focusEditor();
        document.execCommand("insertImage", false, reader.result);
        normalizeImages();
      };
      reader.readAsDataURL(file);
      return; // tek görsel yeter
    }
  }
}

// <img> etiketlerine default stil uygula
function normalizeImages() {
  nextTick(() => {
    const root = editorRef.value;
    if (!root) return;
    root.querySelectorAll("img").forEach((img) => {
      img.style.maxWidth = "100%";
      img.style.height = "auto";
      img.style.display = "block";
      img.style.margin = "8px 0";
      img.alt ||= "image";
      img.referrerPolicy = "no-referrer";
      img.loading = "lazy";
    });
  });
}

function removeFormat() {
  apply("removeFormat");
  apply("unlink"); // linkleri de temizle
}

function undo() {
  apply("undo");
}
function redo() {
  apply("redo");
}

function queryFontSize() {
  const v = document.queryCommandValue("fontSize");
  return v ? Number(v) : 3;
}

function refreshState() {
  nextTick(() => {
    try {
      active.value.bold = document.queryCommandState("bold");
      active.value.italic = document.queryCommandState("italic");
      active.value.underline = document.queryCommandState("underline");
      active.value.strikeThrough = document.queryCommandState("strikeThrough");
      active.value.justifyLeft = document.queryCommandState("justifyLeft");
      active.value.justifyCenter = document.queryCommandState("justifyCenter");
      active.value.justifyRight = document.queryCommandState("justifyRight");
      active.value.justifyFull = document.queryCommandState("justifyFull");
      active.value.unorderedList = document.queryCommandState("insertUnorderedList");
      active.value.orderedList = document.queryCommandState("insertOrderedList");
      fontSizeState.value = queryFontSize();
    } catch {}
  });
}

onMounted(() => {
  selListener = () => refreshState();
  document.addEventListener("selectionchange", selListener, { passive: true });
});
onBeforeUnmount(() => {
  if (selListener) document.removeEventListener("selectionchange", selListener);
});

// ---- Validasyon (hafif, frontend) ----
const error = ref("");
const loading = ref(false);

function validate() {
  const plainBody = (editorRef.value?.innerText || "").trim();
  if (!form.value.title.trim() || form.value.title.trim().length < 3) {
    return "Başlık en az 3 karakter olmalı.";
  }
  if (!plainBody || plainBody.length < 20) {
    return "İçerik en az 20 karakter olmalı.";
  }
  if (!form.value.tags.trim()) {
    return "En az bir etiket girin.";
  }
  return "";
}

async function submit() {
  error.value = "";
  const v = validate();
  if (v) {
    error.value = v;
    return;
  }

  // Plain text gönderiyoruz
  const plainBody = (editorRef.value?.innerText || "").trim();

  try {
    loading.value = true;
    await api.post("/blog", {
      title: form.value.title.trim(),
      body: plainBody,
      type: form.value.type,
      tags: form.value.tags,
      category: form.value.category,
      status: form.value.status,
    });
    alert("Blog oluşturuldu!");
    router.push("/blogs");
  } catch (e) {
    error.value = e?.response?.data?.error || "Oluşturma başarısız";
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <section class="page">
    <div class="header">
      <h1>Blog Oluştur</h1>
      <p>Başlık, içerik ve meta bilgileri girerek yeni bir blog yazısı ekleyin.</p>
    </div>

    <div class="card">
      <div class="grid">
        <div class="field span-2">
          <label for="title">Başlık</label>
          <input id="title" v-model="form.title" :disabled="loading" placeholder="Başlık" />
        </div>

        <!-- Rich Editör -->
        <div class="field span-2">
          <label>İçerik</label>

          <div class="toolbar">
            <!-- Temel -->
            <button type="button" class="tbtn" :class="{ active: active.bold }" @click="apply('bold')" :disabled="loading"><b>B</b></button>
            <button type="button" class="tbtn" :class="{ active: active.italic }" @click="apply('italic')" :disabled="loading"><i>I</i></button>
            <button type="button" class="tbtn" :class="{ active: active.underline }" @click="apply('underline')" :disabled="loading"><u>U</u></button>
            <button type="button" class="tbtn" :class="{ active: active.strikeThrough }" @click="apply('strikeThrough')" :disabled="loading"><s>S</s></button>

            <span class="sep" />

            <!-- Hizalama -->
            <button type="button" class="tbtn" :class="{ active: active.justifyLeft }" @click="setAlign('left')" :disabled="loading" title="Sola yasla">⟸</button>
            <button type="button" class="tbtn" :class="{ active: active.justifyCenter }" @click="setAlign('center')" :disabled="loading" title="Ortala">≡</button>
            <button type="button" class="tbtn" :class="{ active: active.justifyRight }" @click="setAlign('right')" :disabled="loading" title="Sağa yasla">⟹</button>
            <button type="button" class="tbtn" :class="{ active: active.justifyFull }" @click="setAlign('full')" :disabled="loading" title="İki yana yasla">⇔</button>

            <span class="sep" />

            <!-- Liste -->
            <button type="button" class="tbtn" :class="{ active: active.unorderedList }" @click="toggleList('ul')" :disabled="loading">• Liste</button>
            <button type="button" class="tbtn" :class="{ active: active.orderedList }" @click="toggleList('ol')" :disabled="loading">1. Liste</button>

            <span class="sep" />

            <!-- Boyut -->
            <button type="button" class="tbtn" @click="setFontSize(-1)" :disabled="loading">A−</button>
            <div class="tbadge" :title="'Boyut (1..7): ' + fontSizeState">A{{ fontSizeState }}</div>
            <button type="button" class="tbtn" @click="setFontSize(1)" :disabled="loading">A+</button>

            <span class="sep" />

            <!-- Renk + Link + Ayırıcı -->
            <label class="colorPick" title="Yazı rengi">
              <input type="color" :value="foreColor" @input="setColor($event.target.value)" :disabled="loading" />
            </label>
            <button type="button" class="tbtn" @click="insertLink" :disabled="loading">🔗</button>
            <button type="button" class="tbtn" @click="insertSeparator" :disabled="loading">—</button>

            <span class="sep" />

            <!-- Görsel -->
            <button type="button" class="tbtn" @click="insertImageByURL" :disabled="loading">🖼️ URL</button>
            <button type="button" class="tbtn" @click="triggerFilePick" :disabled="loading">🖼️ Dosya</button>
            <input ref="fileInputRef" type="file" accept="image/*" class="hidden-file" @change="onPickFile" />

            <span class="sep" />

            <!-- Undo / Redo / Biçimi temizle -->
            <button type="button" class="tbtn" @click="undo" :disabled="loading">↶</button>
            <button type="button" class="tbtn" @click="redo" :disabled="loading">↷</button>
            <button type="button" class="tbtn" @click="removeFormat" :disabled="loading" title="Biçimi temizle">✖︎</button>
          </div>

          <div
              ref="editorRef"
              class="editor"
              contenteditable="true"
              :data-placeholder="editorPlaceholder"
              :spellcheck="true"
              @keyup="refreshState"
              @mouseup="refreshState"
              @paste="onPaste"
          ></div>

          <p class="hint">
            <strong>Not:</strong> Bu editör görsel biçimlendirme ve görsel eklemeyi destekler;
            içerik <em>düz metin</em> olarak kaydedilir (görsel/biçim kaydedilmez).
          </p>
        </div>

        <div class="field">
          <label for="type">Tip</label>
          <select id="type" v-model="form.type" :disabled="loading">
            <option v-for="t in TYPES" :key="t" :value="t">{{ t }}</option>
          </select>
        </div>

        <div class="field">
          <label for="tags">Etiketler (virgülle)</label>
          <input id="tags" v-model="form.tags" :disabled="loading" placeholder="Örn: go,backend" />
        </div>

        <div class="field">
          <label for="category">Kategori</label>
          <select id="category" v-model="form.category" :disabled="loading">
            <option v-for="c in CATEGORIES" :key="c" :value="c">{{ c }}</option>
          </select>
        </div>

        <div class="field">
          <label for="status">Durum</label>
          <select id="status" v-model="form.status" :disabled="loading">
            <option v-for="s in STATUSES" :key="s" :value="s">{{ s }}</option>
          </select>
        </div>
      </div>

      <div class="actions">
        <button
            class="btn btn-green"
            :disabled="loading"
            @click="submit"
            style="border:1px solid #333;border-radius:6px;padding:6px 10px;cursor:pointer"
        >
          {{ loading ? "Oluşturuluyor…" : "Oluştur" }}
        </button>

        <router-link to="/blogs" class="link">İptal</router-link>
      </div>

      <p v-if="error" class="err">{{ error }}</p>
    </div>
  </section>
</template>

<style scoped>
.page { max-width: 1000px; margin: 0 auto; display: grid; gap: 16px; padding: 8px 0 24px; }
.header h1 { font-size: 28px; font-weight: 800; letter-spacing: .2px; color: var(--color-heading); }
.header p { opacity: .8; margin-top: 2px; font-size: 14px; }

.card {
  border: 1px solid var(--color-border);
  border-radius: 14px;
  padding: 18px;
  background:
      linear-gradient(0deg, rgba(255,255,255,0.02), rgba(255,255,255,0.02)) padding-box,
      radial-gradient(1200px 300px at 0% 0%, rgba(25,210,124,0.12), rgba(25,210,124,0) 60%),
      radial-gradient(1000px 300px at 100% 100%, rgba(25,210,124,0.10), rgba(25,210,124,0) 60%);
  backdrop-filter: blur(6px) saturate(120%);
}

.grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
.span-2 { grid-column: span 2; }

.field { display: grid; gap: 6px; }
.field label { font-size: 13px; font-weight: 700; color: var(--color-heading); letter-spacing: .2px; }
.field input, .field select {
  width: 100%; padding: 10px 12px; border: 1px solid var(--color-border); border-radius: 10px;
  background: var(--color-background); color: var(--color-text); outline: none;
  transition: border-color .15s ease, box-shadow .15s ease;
}
.field input:focus, .field select:focus { border-color: var(--color-border-hover); box-shadow: 0 0 0 2px hsla(160, 100%, 37%, 0.16); }

/* Toolbar + Editor */
.toolbar {
  display: flex; flex-wrap: wrap; gap: 6px; margin-bottom: 8px;
  border: 1px solid var(--color-border); border-radius: 10px; padding: 6px;
  background: var(--color-background-soft);
}
.tbtn {
  border: 1px solid var(--color-border);
  background: var(--color-background);
  color: var(--color-text);
  border-radius: 8px;
  padding: 6px 10px;
  cursor: pointer;
  transition: background .15s ease, border-color .15s ease, opacity .15s ease, box-shadow .15s ease;
}
.tbtn:hover { background: var(--color-background-soft); border-color: var(--color-border-hover); }
.tbtn.active {
  border-color: hsla(160, 100%, 37%, 0.7);
  box-shadow: 0 0 0 2px hsla(160, 100%, 37%, 0.22) inset;
}
.tbadge {
  border: 1px dashed var(--color-border);
  border-radius: 8px;
  padding: 6px 8px;
  font-size: 12px;
  opacity: .9;
}
.sep { width: 1px; height: 24px; background: var(--color-border); margin: 0 4px; }

.colorPick input {
  appearance: none;
  width: 32px; height: 28px;
  padding: 0; border: 1px solid var(--color-border);
  border-radius: 8px; background: transparent; cursor: pointer;
}

.hidden-file { display: none; }

/* Editör kutusu */
.editor {
  min-height: 240px;
  border: 1px solid var(--color-border);
  border-radius: 10px;
  background: var(--color-background);
  color: var(--color-text);
  padding: 10px 12px;
  outline: none;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
.editor:focus { border-color: var(--color-border-hover); box-shadow: 0 0 0 2px hsla(160, 100%, 37%, 0.16); }
.editor:empty:before { content: attr(data-placeholder); opacity: .6; }
.editor img { max-width: 100%; height: auto; display: block; margin: 8px 0; }

/* Alt ipucu */
.hint { font-size: 12px; opacity: .8; margin-top: 6px; }

/* Actions */
.actions { display: flex; align-items: center; gap: 10px; margin-top: 14px; }
.link { text-decoration: none; font-size: 14px; opacity: .9; padding: 6px 10px; border-radius: 8px; border: 1px solid transparent; color: var(--color-text); }
.link:hover { border-color: var(--color-border-hover); background: var(--color-background-soft); }

.err { margin-top: 10px; color: #b00; }

/* Mobil */
@media (max-width: 720px) {
  .grid { grid-template-columns: 1fr; }
  .span-2 { grid-column: span 1; }
}

/* LogNode temalı genel buton stilleri */
.btn { border: 1px solid transparent; border-radius: 8px; padding: 6px 12px; font-weight: 600; cursor: pointer; transition: all 0.2s ease; }
.btn-green { background: hsla(160, 100%, 37%, 1); border-color: hsla(160, 100%, 37%, 1); color: #fff; }
.btn-green:hover { opacity: .9; box-shadow: 0 0 10px rgba(25,210,124,0.3); }
</style>
