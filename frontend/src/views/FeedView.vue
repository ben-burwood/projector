<template>
  <div class="flex flex-col gap-4">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <h1 class="text-2xl font-bold">Feed</h1>
      <button class="btn btn-sm" :disabled="loading" @click="refresh">
        <span v-if="loading" class="loading loading-spinner loading-xs" />
        Refresh
      </button>
    </div>

    <div class="flex flex-wrap items-center gap-2">
      <select v-model="source" class="select select-bordered select-sm" @change="reload">
        <option value="">All sources</option>
        <option v-for="t in SOURCE_TYPES" :key="t" :value="t">{{ sourceInfo(t).label }}</option>
      </select>
      <select v-model="tag" class="select select-bordered select-sm" @change="reload">
        <option value="">All tags</option>
        <option v-for="t in tagsStore.tags" :key="t.id" :value="t.name">{{ t.name }}</option>
      </select>
      <label class="label cursor-pointer gap-1">
        <span class="label-text text-xs">Include not in use</span>
        <input type="checkbox" class="checkbox checkbox-sm" v-model="includeNotInUse" @change="reload" />
      </label>
    </div>

    <!-- Per-project fetch errors, inline, non-fatal -->
    <div v-if="errors.length" class="alert alert-warning flex-col items-start text-sm">
      <span class="font-semibold">Some projects couldn't be fetched:</span>
      <ul class="list-disc pl-5">
        <li v-for="e in errors" :key="e.project_id">{{ e.project_name }}: {{ e.error }}</li>
      </ul>
    </div>

    <div v-if="loading && items.length === 0" class="flex flex-col gap-3">
      <div v-for="n in 4" :key="n" class="skeleton h-28 w-full" />
    </div>

    <div v-else-if="fatalError" class="alert alert-error">{{ fatalError }}</div>

    <EmptyState
      v-else-if="items.length === 0"
      title="Nothing in the feed yet"
      hint="Add some projects with releases (GitHub, PyPI, crates.io…) and they'll show up here."
      icon="📰"
    />

    <div v-else class="flex flex-col gap-3">
      <ReleaseCard v-for="(item, i) in items" :key="`${item.project_id}-${item.release.external_id}-${i}`" :item="item" />
      <button v-if="hasMore" class="btn btn-ghost btn-sm" :disabled="loading" @click="loadMore">
        Load more
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import ReleaseCard from "@/components/ReleaseCard.vue";
import EmptyState from "@/components/EmptyState.vue";
import { api } from "@/lib/api";
import { SOURCE_TYPES, sourceInfo } from "@/lib/sources";
import { useTagsStore } from "@/stores/tags";
import type { FeedError, FeedItem } from "@/lib/types";

const tagsStore = useTagsStore();

const items = ref<FeedItem[]>([]);
const errors = ref<FeedError[]>([]);
const loading = ref(false);
const fatalError = ref("");
const hasMore = ref(false);
const page = ref(1);
const source = ref("");
const tag = ref("");
const includeNotInUse = ref(false);

async function fetchPage(p: number, append: boolean) {
  loading.value = true;
  fatalError.value = "";
  try {
    const res = await api.feed({
      page: p,
      source: source.value,
      tag: tag.value,
      all: includeNotInUse.value || undefined,
    });
    items.value = append ? [...items.value, ...res.items] : res.items;
    errors.value = res.errors ?? [];
    hasMore.value = res.has_more;
    page.value = res.page;
  } catch (e: any) {
    if (!append) items.value = [];
    fatalError.value = e.message ?? String(e);
  } finally {
    loading.value = false;
  }
}

function reload() {
  fetchPage(1, false);
}

function loadMore() {
  fetchPage(page.value + 1, true);
}

async function refresh() {
  try {
    await api.refreshFeed();
  } catch {
    // ignore; still reload from (possibly cached) data
  }
  reload();
}

onMounted(() => {
  tagsStore.load();
  reload();
});
</script>
