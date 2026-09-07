<template>
  <div class="flex flex-col gap-4">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <h1 class="text-2xl font-bold">Projects</h1>
      <RouterLink to="/add" class="btn btn-primary btn-sm">+ Add project</RouterLink>
    </div>

    <!-- Filter bar -->
    <div class="flex flex-wrap items-center gap-2">
      <input
        v-model="filters.q"
        class="input input-bordered input-sm w-48"
        placeholder="Search…"
        @input="debouncedReload"
      />
      <select v-model="filters.source" class="select select-bordered select-sm" @change="reload">
        <option value="">All sources</option>
        <option v-for="t in SOURCE_TYPES" :key="t" :value="t">{{ sourceInfo(t).label }}</option>
      </select>
      <select v-model="filters.tag" class="select select-bordered select-sm" @change="reload">
        <option value="">All tags</option>
        <option v-for="t in tagsStore.tags" :key="t.id" :value="t.name">{{ t.name }}</option>
      </select>
      <label class="label cursor-pointer gap-1">
        <span class="label-text text-xs">In use</span>
        <input type="checkbox" class="checkbox checkbox-sm" v-model="onlyInUse" @change="reload" />
      </label>
    </div>

    <div v-if="store.loading" class="flex justify-center py-16">
      <span class="loading loading-spinner loading-lg" />
    </div>

    <div v-else-if="store.error" class="alert alert-error">{{ store.error }}</div>

    <EmptyState
      v-else-if="store.projects.length === 0"
      title="No projects yet"
      hint="Add a project you like — a GitHub repo, a package, or any website."
      icon="🗂️"
    >
      <RouterLink to="/add" class="btn btn-primary btn-sm">+ Add your first project</RouterLink>
    </EmptyState>

    <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <ProjectCard
        v-for="p in store.projects"
        :key="p.id"
        :project="p"
        @toggle-inuse="store.toggleInUse"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { RouterLink } from "vue-router";
import ProjectCard from "@/components/ProjectCard.vue";
import EmptyState from "@/components/EmptyState.vue";
import { useProjectsStore } from "@/stores/projects";
import { useTagsStore } from "@/stores/tags";
import { SOURCE_TYPES, sourceInfo } from "@/lib/sources";
import type { ProjectFilter } from "@/lib/types";

const store = useProjectsStore();
const tagsStore = useTagsStore();

const filters = reactive<ProjectFilter>({ q: "", source: "", tag: "" });
const onlyInUse = ref(false);

function reload() {
  const f: ProjectFilter = { ...filters };
  if (onlyInUse.value) f.in_use = true;
  store.load(f);
}

let timer: ReturnType<typeof setTimeout> | undefined;
function debouncedReload() {
  clearTimeout(timer);
  timer = setTimeout(reload, 250);
}

onMounted(() => {
  tagsStore.load();
  reload();
});
</script>
