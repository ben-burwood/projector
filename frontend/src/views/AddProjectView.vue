<template>
  <div class="mx-auto flex max-w-xl flex-col gap-4">
    <h1 class="text-2xl font-bold">Add project</h1>

    <!-- Step 1: paste a URL, detect the source -->
    <div class="join w-full">
      <input
        v-model="url"
        class="input input-bordered join-item w-full"
        placeholder="https://github.com/owner/repo, a package, or any website…"
        @keydown.enter.prevent="detect"
      />
      <button class="btn btn-primary join-item" :disabled="detecting || !url.trim()" @click="detect">
        {{ detecting ? "…" : "Detect" }}
      </button>
    </div>

    <div v-if="detectError" class="alert alert-warning text-sm">
      {{ detectError }} — you can still save it as a plain website bookmark below.
    </div>

    <!-- Step 2: confirm/edit prefilled details -->
    <div v-if="form" class="card bg-base-100 shadow-md">
      <div class="card-body gap-3">
        <div class="flex items-center gap-2">
          <span class="text-sm font-semibold">Detected as</span>
          <SourceBadge :type="form.source_type" />
          <span v-if="form.source_identifier" class="text-base-content/60 font-mono text-xs">
            {{ form.source_identifier }}
          </span>
        </div>

        <label class="form-control">
          <span class="label-text">Name</span>
          <input v-model="form.name" class="input input-bordered input-sm" />
        </label>

        <div class="form-control">
          <span class="label-text mb-1">Tags</span>
          <TagPicker v-model="tagIds" />
        </div>

        <label class="label cursor-pointer justify-start gap-2">
          <input type="checkbox" class="toggle toggle-sm toggle-primary" v-model="inUse" />
          <span class="label-text">In use</span>
        </label>

        <div v-if="saveError" class="alert alert-error text-sm">{{ saveError }}</div>

        <div class="card-actions justify-end">
          <button class="btn btn-ghost btn-sm" @click="reset">Cancel</button>
          <button class="btn btn-primary btn-sm" :disabled="saving" @click="save">
            {{ saving ? "Saving…" : "Save project" }}
          </button>
        </div>
      </div>
    </div>

    <div class="divider">or</div>

    <!-- Bulk import: every repo a GitHub user has starred -->
    <div class="card bg-base-100 shadow-md">
      <div class="card-body gap-3">
        <h2 class="card-title text-lg">Import GitHub stars</h2>

        <input
          v-model="starUser"
          class="input input-bordered input-sm w-full"
          placeholder="GitHub username (blank = token owner)"
          @keydown.enter.prevent="importStars"
        />

        <div class="form-control">
          <span class="label-text mb-1">Tag imported projects</span>
          <TagPicker v-model="starTagIds" />
        </div>

        <label class="label cursor-pointer justify-start gap-2">
          <input type="checkbox" class="toggle toggle-sm toggle-primary" v-model="starInUse" />
          <span class="label-text">Mark as in use</span>
        </label>

        <div v-if="importError" class="alert alert-error text-sm">{{ importError }}</div>
        <div v-if="importResult" class="alert alert-success text-sm">
          Imported {{ importResult.imported }} of {{ importResult.total }} starred repos
          <span v-if="importResult.skipped">({{ importResult.skipped }} already tracked)</span>.
          <RouterLink to="/projects" class="link">View projects</RouterLink>
        </div>

        <div class="card-actions justify-end">
          <button class="btn btn-primary btn-sm" :disabled="importing" @click="importStars">
            {{ importing ? "Importing…" : "Import stars" }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { RouterLink, useRouter } from "vue-router";
import SourceBadge from "@/components/SourceBadge.vue";
import TagPicker from "@/components/TagPicker.vue";
import { api } from "@/lib/api";
import type { DetectResponse, ImportResult, SourceType } from "@/lib/types";

const router = useRouter();

const url = ref("");
const detecting = ref(false);
const detectError = ref("");

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
