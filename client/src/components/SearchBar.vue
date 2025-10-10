<script setup>
import { ref, watch } from "vue";

const props = defineProps({
  modelValue: { type: String, default: "" },
  mode: { type: String, default: "user" }, // "user" | "blog"
});
const emits = defineEmits(["update:modelValue", "update:mode", "search"]);

const keyword = ref(props.modelValue);
const activeMode = ref(props.mode);
let t = null;

watch(keyword, (v) => {
  emits("update:modelValue", v);
  clearTimeout(t);
  t = setTimeout(() => emits("search", { q: v, mode: activeMode.value }), 300);
});

function setMode(m) {
  activeMode.value = m;
  emits("update:mode", m);
  emits("search", { q: keyword.value, mode: m });
}
//  <div style="border:1px solid #ddd;border-radius:8px;overflow:hidden;display:flex">
//  <button :class="{on: activeMode==='user'}" @click="setMode('user')">Kullanıcı</button>
//  <button :class="{on: activeMode==='blog'}" @click="setMode('blog')">Blog</button>   Kullanıcı butonunu altındaydı
//  </div>
</script>

<template>
  <div style="display:flex;gap:8px;align-items:center">
    <input
        :value="keyword"
        @input="keyword = $event.target.value"
        placeholder="Ara (ör. d, de, den...)"
        style="flex:1;padding:8px;border:1px solid #ddd;border-radius:8px"
    />
  </div>
</template>

<style>
button.on { background:#111; color:#fff; }
button { padding:6px 10px; border:none; background:#f5f5f5; cursor:pointer }
button + button { border-left:1px solid #ddd; }
</style>
