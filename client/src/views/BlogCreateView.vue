<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import api from "../api/axios";

const router = useRouter();

const form = ref({
  title: "",
  body: "",
  type: "",      // ör: "tech" / "life" vs — backend zorunlu tuttuğu için boş geçmeyelim
  tags: "",      // virgülle ayırabilirsin: "go,backend"
  category: "",  // ör: "golang"
  status: "",    // ör: "draft" / "published" vs — backend zorunlu
});

const error = ref("");
const loading = ref(false);

async function submit() {
  error.value = "";
  if (!form.value.title || !form.value.body || !form.value.type || !form.value.status || !form.value.tags || !form.value.category) {
    error.value = "Tüm alanları doldurun.";
    return;
  }

  try {
    loading.value = true;
    await api.post("/blog", {
      title: form.value.title,
      body: form.value.body,
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
  <section style="max-width:700px;margin:auto;display:grid;gap:16px">
    <h1>Blog Oluştur</h1>

    <div style="display:grid;gap:8px;border:1px solid #eee;padding:12px;border-radius:8px">
      <label>Başlık</label>
      <input v-model="form.title" :disabled="loading" placeholder="Başlık" />

      <label>İçerik</label>
      <textarea v-model="form.body" :disabled="loading" rows="8" placeholder="İçerik"></textarea>

      <label>Tip</label>
      <input v-model="form.type" :disabled="loading" placeholder="Örn: tech" />

      <label>Etiketler (virgülle)</label>
      <input v-model="form.tags" :disabled="loading" placeholder="Örn: go,backend" />

      <label>Kategori</label>
      <input v-model="form.category" :disabled="loading" placeholder="Örn: golang" />

      <label>Durum</label>
      <input v-model="form.status" :disabled="loading" placeholder="Örn: draft/published" />

      <div style="display:flex;gap:8px;margin-top:8px">
        <button @click="submit" :disabled="loading">Oluştur</button>
        <router-link to="/blogs">İptal</router-link>
      </div>

      <p v-if="error" style="color:#b00">{{ error }}</p>
    </div>
  </section>
</template>
