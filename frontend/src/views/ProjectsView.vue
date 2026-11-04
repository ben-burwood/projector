<template>
  <div>
    <PageHeader title="Projects">
      <template #subtitle>
        <template v-if="store.projects.length || !store.loading">
          {{ store.projects.length }} project{{ store.projects.length === 1 ? "" : "s" }}
          <template v-if="inUseCount">· {{ inUseCount }} in use</template>
        </template>
      </template>
      <RouterLink to="/add" class="btn btn-primary btn-sm">
        <Icon name="plus" /> Add project
      </RouterLink>
    </PageHeader>

    <!-- Filter bar -->
    <div class="mb-5 flex flex-wrap items-center gap-2">
      <div class="relative w-full sm:w-64">
        <Icon name="search" class="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-muted" />
        <input
          v-model="filters.q"
          class="field pl-9"
          placeholder="Search projects…"
          @input="debouncedReload"
        />
      </div>
      <select v-model="filters.source" class="field w-auto" @change="reload">
        <option value="">All sources</option>
        <option v-for="t in SOURCE_TYPES" :key="t" :value="t">{{ sourceInfo(t).label }}</option>
      </select>
      <select v-model="filters.tag" class="field w-auto" @change="reload">
        <option value="">All tags</option>
        <option v-for="t in tagsStore.tags" :key="t.id" :value="t.name">#{{ t.name }}</option>
      </select>

      <div class="flex h-9 items-center rounded-lg border border-line bg-surface p-0.5 text-xs">
        <button
          v-for="opt in USE_OPTIONS"
          :key="opt.label"
          type="button"
          class="h-full cursor-pointer rounded-md px-2.5 transition-colors"
          :class="useFilter === opt.value ? 'bg-surface-2 font-medium text-ink' : 'text-muted hover:text-ink'"
          @click="setUseFilter(opt.value)"
        >
          {{ opt.label }}
        </button>
      </div>
    </div>

    <div v-if="store.loading && store.projects.length === 0" class="panel divide-y divide-line">
      <div v-for="n in 6" :key="n" class="flex items-center gap-3 px-4 py-3">
        <div class="skeleton size-10" />
        <div class="flex flex-1 flex-col gap-2">
          <div class="skeleton h-3.5 w-40" />
          <div class="skeleton h-3 w-24" />
        </div>
      </div>
    </div>

    <div v-else-if="store.error" class="notice notice-error">
      <Icon name="alert" class="mt-0.5" /> {{ store.error }}
    </div>

    <EmptyState
      v-else-if="store.projects.length === 0 && !hasFilters"
      title="No projects yet"
      hint="Add a project you like — a GitHub repo, a package, or any website. Or import your GitHub stars."
      icon="folder"
    >
      <RouterLink to="/add" class="btn btn-primary btn-sm"><Icon name="plus" /> Add your first project</RouterLink>
    </EmptyState>

    <EmptyState
      v-else-if="store.projects.length === 0"
      title="No matches"
      hint="Nothing matches these filters."
      icon="search"
    >
      <button class="btn btn-outline btn-sm" @click="clearFilters">Clear filters</button>
    </EmptyState>

    <div v-else class="panel divide-y divide-line overflow-hidden" :class="store.loading && 'opacity-60'">
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
import { computed, onMounted, reactive, ref } from "vue";
import { RouterLink, useRoute } from "vue-router";
import EmptyState from "@/components/EmptyState.vue";
import Icon from "@/components/Icon.vue";
import PageHeader from "@/components/PageHeader.vue";
import ProjectCard from "@/components/ProjectCard.vue";
import { useProjectsStore } from "@/stores/projects";
import { useTagsStore } from "@/stores/tags";
import { SOURCE_TYPES, sourceInfo } from "@/lib/sources";
import type { ProjectFilter } from "@/lib/types";

const store = useProjectsStore();
const tagsStore = useTagsStore();
const route = useRoute();

const USE_OPTIONS: { label: string; value: boolean | undefined }[] = [
  { label: "All", value: undefined },
  { label: "In use", value: true },
  { label: "Not in use", value: false },
];

// The Tags page links here with ?tag=name.
const filters = reactive<ProjectFilter>({
  q: "",
  source: "",
  tag: typeof route.query.tag === "string" ? route.query.tag : "",
});
const useFilter = ref<boolean | undefined>(undefined);

const inUseCount = computed(() => store.projects.filter((p) => p.in_use).length);
const hasFilters = computed(
  () => !!(filters.q || filters.source || filters.tag) || useFilter.value !== undefined,
);

function reload() {
  store.load({ ...filters, in_use: useFilter.value });
}

function setUseFilter(v: boolean | undefined) {
  useFilter.value = v;
  reload();
}

function clearFilters() {
  Object.assign(filters, { q: "", source: "", tag: "" });
  useFilter.value = undefined;
  reload();
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
