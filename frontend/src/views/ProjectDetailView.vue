<template>
  <div v-if="loading" class="flex justify-center py-16">
    <span class="loading loading-spinner loading-lg" />
  </div>

  <div v-else-if="loadError" class="alert alert-error">{{ loadError }}</div>

  <div v-else-if="project" class="flex flex-col gap-6">
    <!-- Header -->
    <div class="flex flex-wrap items-start justify-between gap-2">
      <div class="flex flex-col gap-1">
        <h1 class="text-2xl font-bold">{{ project.name }}</h1>
        <a :href="project.url" target="_blank" rel="noopener" class="link link-primary text-sm">
          {{ project.url }} ↗
        </a>
      </div>
      <div class="flex gap-2">
        <button class="btn btn-ghost btn-sm" @click="editing = true">Edit</button>
        <button class="btn btn-ghost btn-sm text-error" @click="remove">Delete</button>
      </div>
    </div>

    <div class="flex flex-wrap items-center gap-2">
      <SourceBadge :type="project.source_type" />
      <span v-if="project.source_identifier" class="text-base-content/60 font-mono text-xs">
        {{ project.source_identifier }}
      </span>
      <label class="label ml-2 cursor-pointer gap-2 py-0">
        <span class="label-text text-xs">In use</span>
        <input type="checkbox" class="toggle toggle-xs toggle-primary" :checked="project.in_use" @change="toggleInUse" />
      </label>
    </div>

    <div class="flex flex-wrap items-center gap-1">
      <TagChip v-for="t in project.tags" :key="t.id" :tag="t" />
    </div>

    <!-- Live release history -->
    <div class="flex flex-col gap-3">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-semibold">Releases</h2>
        <button
          v-if="pollable"
          class="btn btn-xs"
          :disabled="releasesLoading"
          @click="refreshReleases"
        >
          <span v-if="releasesLoading" class="loading loading-spinner loading-xs" />
          Refresh
        </button>
      </div>

      <div v-if="!pollable" class="text-base-content/60 text-sm">
        This is a plain website bookmark — no releases are tracked.
      </div>

      <template v-else>
        <div v-if="releasesLoading && releases.length === 0" class="flex flex-col gap-2">
          <div v-for="n in 3" :key="n" class="skeleton h-20 w-full" />
        </div>
        <div v-else-if="releasesError" class="alert alert-warning text-sm">{{ releasesError }}</div>
        <EmptyState v-else-if="releases.length === 0" title="No releases found" icon="🏷️" />
        <ul v-else class="flex flex-col gap-3">
          <li v-for="(r, i) in releases" :key="`${r.external_id}-${i}`" class="card bg-base-100 shadow-sm">
            <div class="card-body gap-1 p-4">
              <div class="flex items-center justify-between gap-2">
                <span class="badge badge-primary badge-sm font-mono">{{ r.version || r.title }}</span>
                <span class="text-base-content/50 text-xs">{{ fmtDate(r.published_at) }}</span>
              </div>
              <p v-if="r.body" class="text-base-content/70 whitespace-pre-line text-sm">
                {{ r.body.length > 400 ? r.body.slice(0, 400) + "…" : r.body }}
              </p>
              <a v-if="r.url" :href="r.url" target="_blank" rel="noopener" class="link link-primary text-xs">
                View ↗
              </a>
            </div>
          </li>
        </ul>
      </template>
    </div>

    <!-- Edit modal -->
    <dialog class="modal" :open="editing">
      <div class="modal-box">
        <h3 class="text-lg font-bold">Edit project</h3>
        <div class="flex flex-col gap-3 py-3">
          <label class="form-control">
            <span class="label-text">Name</span>
            <input v-model="editForm.name" class="input input-bordered input-sm" />
          </label>
          <label class="form-control">
            <span class="label-text">URL</span>
            <input v-model="editForm.url" class="input input-bordered input-sm" />
          </label>
          <div class="form-control">
            <span class="label-text mb-1">Tags</span>
            <TagPicker v-model="editTagIds" />
          </div>
        </div>
        <div class="modal-action">
          <button class="btn btn-ghost btn-sm" @click="editing = false">Cancel</button>
          <button class="btn btn-primary btn-sm" @click="saveEdit">Save</button>
        </div>
      </div>
      <div class="modal-backdrop" @click="editing = false" />
    </dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import SourceBadge from "@/components/SourceBadge.vue";
import TagChip from "@/components/TagChip.vue";
import TagPicker from "@/components/TagPicker.vue";
import EmptyState from "@/components/EmptyState.vue";
import { api } from "@/lib/api";
import { sourceInfo } from "@/lib/sources";
import type { Project, Release } from "@/lib/types";

const route = useRoute();
const router = useRouter();

const project = ref<Project | null>(null);
const loading = ref(false);
const loadError = ref("");

const releases = ref<Release[]>([]);
const releasesLoading = ref(false);
const releasesError = ref("");

const editing = ref(false);
const editForm = reactive({ name: "", url: "" });
const editTagIds = ref<number[]>([]);

const pollable = computed(() =>
  project.value ? sourceInfo(project.value.source_type).pollable : false,
);

function projectId(): number {
  return Number(route.params.id);
}

async function load() {
  loading.value = true;
  loadError.value = "";
  try {
    project.value = await api.getProject(projectId());
    if (pollable.value) loadReleases();
  } catch (e: any) {
    loadError.value = e.message ?? String(e);
  } finally {
    loading.value = false;
  }
}

async function loadReleases() {
  releasesLoading.value = true;
  releasesError.value = "";
  try {
    releases.value = await api.projectReleases(projectId());
  } catch (e: any) {
    releasesError.value = e.message ?? String(e);
  } finally {
    releasesLoading.value = false;
  }
}

async function refreshReleases() {
  try {
    await api.refreshProject(projectId());
  } catch {
    /* ignore */
  }
  loadReleases();
}

async function toggleInUse() {
  if (!project.value) return;
  project.value = await api.updateProject(project.value.id, { in_use: !project.value.in_use });
}

async function remove() {
  if (!project.value) return;
  if (confirm("Delete this project?")) {
    await api.deleteProject(project.value.id);
    router.push("/projects");
  }
}

watch(editing, (open) => {
  if (open && project.value) {
    editForm.name = project.value.name;
    editForm.url = project.value.url;
    editTagIds.value = project.value.tags.map((t) => t.id);
  }
});

async function saveEdit() {
  if (!project.value) return;
  project.value = await api.updateProject(project.value.id, {
    name: editForm.name,
    url: editForm.url,
    tag_ids: editTagIds.value,
  });
  editing.value = false;
}

function fmtDate(s: string): string {
  if (!s) return "";
  const d = new Date(s);
  return isNaN(d.getTime()) ? "" : d.toLocaleDateString();
}

onMounted(load);
watch(() => route.params.id, load);
</script>
