import { defineStore } from "pinia";
import { ref } from "vue";
import { api } from "@/lib/api";
import type { Project, ProjectFilter } from "@/lib/types";

export const useProjectsStore = defineStore("projects", () => {
  const projects = ref<Project[]>([]);
  const loading = ref(false);
  const error = ref("");

  async function load(filter: ProjectFilter = {}) {
    loading.value = true;
    error.value = "";
    try {
      projects.value = await api.listProjects(filter);
    } catch (e: any) {
      error.value = e.message ?? String(e);
    } finally {
      loading.value = false;
    }
  }

  // Optimistically update a single project in the local list.
  function upsertLocal(p: Project) {
    const i = projects.value.findIndex((x) => x.id === p.id);
    if (i >= 0) projects.value[i] = p;
    else projects.value.unshift(p);
  }

  async function toggleInUse(p: Project) {
    const updated = await api.updateProject(p.id, { in_use: !p.in_use });
    upsertLocal(updated);
    return updated;
  }

  async function remove(id: number) {
    await api.deleteProject(id);
    projects.value = projects.value.filter((p) => p.id !== id);
  }

  return {
    projects,
    loading,
    error,
    load,
    upsertLocal,
    toggleInUse,
    remove,
  };
});
