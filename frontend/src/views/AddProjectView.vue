<template>
  <div class="mx-auto max-w-xl">
    <PageHeader title="Add project" subtitle="Paste a link — Projector works out the source and fills in the details." />

    <div class="flex flex-col gap-4">
      <!-- Step 1: paste a URL, detect the source -->
      <form class="panel flex flex-col gap-4 p-5" @submit.prevent="detect">
        <div class="flex gap-2">
          <div class="relative flex-1">
            <Icon name="link" class="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-muted" />
            <input
              v-model="url"
              class="field h-10 pl-9"
              placeholder="github.com/owner/repo, a package, or any website…"
              autofocus
            />
          </div>
          <button type="submit" class="btn btn-primary h-10" :disabled="detecting || !url.trim()">
            <span v-if="detecting" class="spinner" />
            Detect
          </button>
        </div>

        <div class="flex flex-wrap items-center gap-x-3 gap-y-1.5 text-xs text-muted">
          <span>Supports</span>
          <SourceBadge v-for="t in SOURCE_TYPES" :key="t" :type="t" />
        </div>

        <div v-if="detectError" class="notice notice-warn">
          <Icon name="alert" class="mt-0.5" />
          <span>{{ detectError }} — you can still save it as a plain website bookmark.</span>
        </div>

        <!-- Step 2: confirm/edit prefilled details -->
        <div v-if="form" class="flex flex-col gap-4 border-t border-line pt-4">
          <div class="flex items-center gap-3">
            <ProjectAvatar
              :name="form.name"
              :type="form.source_type"
              :identifier="form.source_identifier"
            />
            <div class="flex min-w-0 flex-col gap-0.5">
              <SourceBadge :type="form.source_type" />
              <span v-if="form.source_identifier" class="truncate font-mono text-xs text-muted">
                {{ form.source_identifier }}
              </span>
            </div>
          </div>

          <p v-if="description" class="text-sm text-muted">{{ description }}</p>

          <label class="flex flex-col gap-1.5">
            <span class="eyebrow">Name</span>
            <input v-model="form.name" class="field" />
          </label>

          <div class="flex flex-col gap-1.5">
            <span class="eyebrow">Tags</span>
            <TagPicker v-model="tagIds" />
          </div>

          <Toggle v-model="inUse" label="In use — show its releases in the feed" />

          <div v-if="saveError" class="notice notice-error">
            <Icon name="alert" class="mt-0.5" /> {{ saveError }}
          </div>

          <div class="flex justify-end gap-2">
            <button type="button" class="btn btn-ghost btn-sm" @click="reset">Cancel</button>
            <button type="button" class="btn btn-primary btn-sm" :disabled="saving" @click="save">
              <span v-if="saving" class="spinner" />
              Save project
            </button>
          </div>
        </div>
      </form>

      <div class="flex items-center gap-3 text-xs text-muted">
        <span class="h-px flex-1 bg-line" /> or <span class="h-px flex-1 bg-line" />
      </div>

      <!-- Bulk import: every repo a GitHub user has starred -->
      <form class="panel flex flex-col gap-4 p-5" @submit.prevent="importStars">
        <div class="flex items-start gap-3">
          <div class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-accent/15 text-accent-text">
            <Icon name="star" class="size-5" />
          </div>
          <div class="flex flex-col gap-0.5">
            <h2 class="font-semibold">Import GitHub stars</h2>
            <p class="text-sm text-muted">
              Bookmark every repo a user has starred. Ones you already track are skipped.
            </p>
          </div>
        </div>

        <div class="relative">
          <span class="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-sm text-muted">@</span>
          <input
            v-model="starUser"
            class="field pl-8"
            placeholder="username (blank = owner of the configured token)"
          />
        </div>

        <div class="flex flex-col gap-1.5">
          <span class="eyebrow">Tag imported projects</span>
          <TagPicker v-model="starTagIds" />
        </div>

        <Toggle v-model="starInUse" label="Mark as in use" />

        <div v-if="importError" class="notice notice-error">
          <Icon name="alert" class="mt-0.5" /> {{ importError }}
        </div>
        <div v-if="importResult" class="notice notice-ok">
          <Icon name="check" class="mt-0.5" />
          <span>
            Imported <strong>{{ importResult.imported }}</strong> of {{ importResult.total }} starred repos<template
              v-if="importResult.skipped"
            >
              ({{ importResult.skipped }} already tracked)</template>.
            <RouterLink to="/projects" class="font-medium underline underline-offset-2">View projects</RouterLink>
          </span>
        </div>

        <div class="flex justify-end">
          <button type="submit" class="btn btn-primary btn-sm" :disabled="importing">
            <span v-if="importing" class="spinner" />
            <Icon v-else name="download" />
            {{ importing ? "Importing…" : "Import stars" }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { RouterLink, useRouter } from "vue-router";
import Icon from "@/components/Icon.vue";
import PageHeader from "@/components/PageHeader.vue";
import ProjectAvatar from "@/components/ProjectAvatar.vue";
import SourceBadge from "@/components/SourceBadge.vue";
import TagPicker from "@/components/TagPicker.vue";
import Toggle from "@/components/Toggle.vue";
import { api } from "@/lib/api";
import { SOURCE_TYPES } from "@/lib/sources";
import type { DetectResponse, ImportResult, SourceType } from "@/lib/types";

const router = useRouter();

const url = ref("");
const detecting = ref(false);
const detectError = ref("");
const description = ref("");

interface Form {
  name: string;
  source_type: SourceType;
  source_identifier: string;
  source_host: string;
}
const form = ref<Form | null>(null);
const tagIds = ref<number[]>([]);
const inUse = ref(true);

const saving = ref(false);
const saveError = ref("");

function applyDetect(d: DetectResponse) {
  description.value = d.prefill.description;
  form.value = {
    name: d.prefill.name || url.value,
    source_type: d.source_type,
    source_identifier: d.source_identifier,
    source_host: d.source_host,
  };
}

async function detect() {
  const u = url.value.trim();
  if (!u) return;
  detecting.value = true;
  detectError.value = "";
  description.value = "";
  form.value = null;
  try {
    const d = await api.detect(u);
    applyDetect(d);
  } catch (e: any) {
    // Detection endpoint not available yet, or unrecognised URL: fall back to website.
    detectError.value = e.message ?? String(e);
    form.value = {
      name: u,
      source_type: "website",
      source_identifier: "",
      source_host: "",
    };
  } finally {
    detecting.value = false;
  }
}

async function save() {
  if (!form.value) return;
  saving.value = true;
  saveError.value = "";
  try {
    const created = await api.createProject({
      url: url.value.trim(),
      name: form.value.name,
      source_type: form.value.source_type,
      source_identifier: form.value.source_identifier,
      source_host: form.value.source_host,
      in_use: inUse.value,
      tag_ids: tagIds.value,
    });
    router.push(`/projects/${created.id}`);
  } catch (e: any) {
    saveError.value = e.message ?? String(e);
  } finally {
    saving.value = false;
  }
}

function reset() {
  form.value = null;
  url.value = "";
  tagIds.value = [];
  detectError.value = "";
  description.value = "";
}

const starUser = ref("");
const starTagIds = ref<number[]>([]);
const starInUse = ref(false);
const importing = ref(false);
const importError = ref("");
const importResult = ref<ImportResult | null>(null);

async function importStars() {
  importing.value = true;
  importError.value = "";
  importResult.value = null;
  try {
    importResult.value = await api.importGitHubStars({
      username: starUser.value.trim(),
      in_use: starInUse.value,
      tag_ids: starTagIds.value,
    });
  } catch (e: any) {
    importError.value = e.message ?? String(e);
  } finally {
    importing.value = false;
  }
}
</script>
