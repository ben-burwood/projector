import { defineStore } from "pinia";
import { ref } from "vue";
import { api } from "@/lib/api";
import type { Tag } from "@/lib/types";

export const useTagsStore = defineStore("tags", () => {
  const tags = ref<Tag[]>([]);
  const loaded = ref(false);
  const loading = ref(false);
  const error = ref("");

  async function load(force = false) {
    if (loaded.value && !force) return;
    loading.value = true;
    error.value = "";
    try {
      tags.value = await api.listTags();
      loaded.value = true;
    } catch (e: any) {
      error.value = e.message ?? String(e);
    } finally {
      loading.value = false;
    }
  }

  async function create(name: string) {
    const tag = await api.createTag(name);
    await load(true);
    return tag;
  }

  async function rename(id: number, name: string) {
    await api.renameTag(id, name);
    await load(true);
  }

  async function remove(id: number) {
    await api.deleteTag(id);
    await load(true);
  }

  return { tags, loaded, loading, error, load, create, rename, remove };
});
