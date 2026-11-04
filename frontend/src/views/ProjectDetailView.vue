<template>
  <div>
    <RouterLink to="/projects" class="mb-5 inline-flex items-center gap-1.5 text-sm text-muted hover:text-ink">
      <Icon name="back" /> Projects
    </RouterLink>

    <div v-if="loading" class="flex flex-col gap-4">
      <div class="flex items-center gap-4">
        <div class="skeleton size-14" />
        <div class="flex flex-col gap-2">
          <div class="skeleton h-6 w-48" />
          <div class="skeleton h-4 w-32" />
        </div>
      </div>
      <div class="skeleton h-64" />
    </div>

    <div v-else-if="loadError" class="notice notice-error">
      <Icon name="alert" class="mt-0.5" /> {{ loadError }}
    </div>

    <div v-else-if="project" class="flex flex-col gap-8">
      <!-- Header -->
      <header class="flex flex-col gap-4">
        <div class="flex flex-wrap items-start gap-4">
          <ProjectAvatar
            :name="project.name"
            :type="project.source_type"
            :identifier="project.source_identifier"
            size="lg"
          />
          <div class="flex min-w-0 flex-1 flex-col gap-1">
            <h1 class="text-2xl font-semibold tracking-tight break-words">{{ project.name }}</h1>
            <div class="flex flex-wrap items-center gap-x-3 gap-y-1">
              <SourceBadge :type="project.source_type" />
              <span v-if="project.source_identifier" class="font-mono text-xs text-muted">
                {{ project.source_identifier }}
              </span>
            </div>
          </div>
          <div class="flex items-center gap-1">
            <a :href="project.url" target="_blank" rel="noopener" class="btn btn-outline btn-sm">
              Open <Icon name="external" class="size-3.5" />
            </a>
            <button class="btn btn-ghost btn-sm btn-icon" aria-label="Edit" title="Edit" @click="editing = true">
              <Icon name="pencil" />
            </button>
            <button class="btn btn-danger btn-sm btn-icon" aria-label="Delete" title="Delete" @click="remove">
              <Icon name="trash" />
            </button>
          </div>
        </div>

        <div class="panel flex flex-wrap items-center gap-x-6 gap-y-3 px-4 py-3">
          <Toggle :model-value="project.in_use" label="In use" @update:model-value="toggleInUse" />
          <div class="h-5 w-px bg-line max-sm:hidden" />
          <div class="flex min-w-0 flex-1 flex-wrap items-center gap-1.5">
            <TagChip v-for="t in project.tags" :key="t.id" :tag="t" removable @remove="removeTag" />
            <button class="chip cursor-pointer border-dashed hover:border-muted/50 hover:text-ink" @click="editing = true">
              <Icon name="plus" class="size-3" /> Tag
            </button>
          </div>
          <span class="text-xs text-muted">Added {{ fmtDate(project.created_at) }}</span>
        </div>
      </header>

      <!-- Live release history -->
      <section class="flex flex-col gap-4">
        <div class="flex items-center justify-between">
          <h2 class="eyebrow">Releases</h2>
          <button
            v-if="pollable"
            class="btn btn-ghost btn-sm"
            :disabled="releasesLoading"
            @click="refreshReleases"
          >
            <Icon name="refresh" :class="releasesLoading && 'animate-spin'" />
            Refresh
          </button>
        </div>

        <EmptyState
          v-if="!pollable"
          title="Bookmark only"
          hint="This is a plain website bookmark — there are no releases to track."
          icon="link"
        />

        <template v-else>
          <div v-if="releasesLoading && releases.length === 0" class="flex flex-col gap-3">
            <div v-for="n in 3" :key="n" class="skeleton h-24" />
          </div>
          <div v-else-if="releasesError" class="notice notice-warn">
            <Icon name="alert" class="mt-0.5" /> {{ releasesError }}
          </div>
          <EmptyState v-else-if="releases.length === 0" title="No releases found" icon="tag" />

          <ol v-else class="relative flex flex-col">
            <li
              v-for="(r, i) in releases"
              :key="`${r.external_id}-${i}`"
              class="relative flex gap-4 pb-6 last:pb-0"
            >
              <!-- timeline rail -->
              <div class="relative flex w-3 shrink-0 justify-center">
                <span
                  class="relative z-10 mt-1.5 size-3 rounded-full border-2"
                  :class="i === 0 ? 'border-accent bg-accent' : 'border-line bg-bg'"
                />
                <span v-if="i < releases.length - 1" class="absolute top-4 -bottom-1.5 w-px bg-line" />
              </div>

              <div class="flex min-w-0 flex-1 flex-col gap-1.5">
                <div class="flex flex-wrap items-center gap-2">
                  <a
                    v-if="r.url"
                    :href="r.url"
                    target="_blank"
                    rel="noopener"
                    class="version hover:bg-accent/25"
                  >
                    {{ r.version || r.title }}
                  </a>
                  <span v-else class="version">{{ r.version || r.title }}</span>
                  <span v-if="i === 0" class="chip h-5 border-ok/40 text-[0.65rem] font-medium text-ok">Latest</span>
                  <span
                    v-if="r.title && r.title !== r.version"
                    class="min-w-0 truncate text-sm font-medium"
                  >
                    {{ r.title }}
                  </span>
                  <span class="ml-auto text-xs text-muted" :title="fmtDate(r.published_at)">
                    {{ relativeTime(r.published_at) }}
                  </span>
                </div>
                <p
                  v-if="plainText(r.body)"
                  class="text-sm whitespace-pre-line text-muted"
                  :class="expanded.has(i) ? '' : 'line-clamp-4'"
                >
                  {{ plainText(r.body) }}
                </p>
                <button
                  v-if="plainText(r.body).length > 300"
                  type="button"
                  class="w-fit cursor-pointer text-xs text-muted hover:text-ink"
                  @click="toggleExpanded(i)"
                >
                  {{ expanded.has(i) ? "Show less" : "Show more" }}
                </button>
              </div>
            </li>
          </ol>
        </template>
      </section>

      <Modal v-model:open="editing" title="Edit project">
        <div class="flex flex-col gap-4">
          <label class="flex flex-col gap-1.5">
            <span class="eyebrow">Name</span>
            <input v-model="editForm.name" class="field" />
          </label>
          <label class="flex flex-col gap-1.5">
            <span class="eyebrow">URL</span>
            <input v-model="editForm.url" class="field" />
          </label>
          <div class="flex flex-col gap-1.5">
            <span class="eyebrow">Tags</span>
            <TagPicker v-model="editTagIds" />
          </div>
        </div>
        <template #actions>
          <button class="btn btn-ghost btn-sm" @click="editing = false">Cancel</button>
          <button class="btn btn-primary btn-sm" @click="saveEdit">Save</button>
        </template>
      </Modal>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import EmptyState from "@/components/EmptyState.vue";
import Icon from "@/components/Icon.vue";
import Modal from "@/components/Modal.vue";
import ProjectAvatar from "@/components/ProjectAvatar.vue";
import SourceBadge from "@/components/SourceBadge.vue";
import TagChip from "@/components/TagChip.vue";
import TagPicker from "@/components/TagPicker.vue";
import Toggle from "@/components/Toggle.vue";
import { api } from "@/lib/api";
import { fmtDate, plainText, relativeTime } from "@/lib/format";
import { sourceInfo } from "@/lib/sources";
import type { Project, Release, Tag } from "@/lib/types";

const route = useRoute();
const router = useRouter();

const project = ref<Project | null>(null);
const loading = ref(false);
const loadError = ref("");

const releases = ref<Release[]>([]);
const releasesLoading = ref(false);
const releasesError = ref("");
const expanded = reactive(new Set<number>());

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
  releases.value = [];
  expanded.clear();
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

function toggleExpanded(i: number) {
  expanded.has(i) ? expanded.delete(i) : expanded.add(i);
}

async function toggleInUse() {
  if (!project.value) return;
  project.value = await api.updateProject(project.value.id, { in_use: !project.value.in_use });
}

async function removeTag(tag: Tag) {
  if (!project.value) return;
  project.value = await api.setProjectTags(
    project.value.id,
    project.value.tags.filter((t) => t.id !== tag.id).map((t) => t.id),
  );
}

async function remove() {
  if (!project.value) return;
  if (confirm(`Delete "${project.value.name}"?`)) {
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

onMounted(load);
watch(() => route.params.id, load);
</script>
