<template>
  <div>
    <PageHeader title="Feed" subtitle="Latest releases across everything you follow, fetched live.">
      <button class="btn btn-outline btn-sm" :disabled="loading" @click="refresh">
        <Icon name="refresh" :class="loading && 'animate-spin'" />
        Refresh
      </button>
    </PageHeader>

    <div class="mb-5 flex flex-wrap items-center gap-2">
      <select v-model="source" class="field w-auto" @change="reload">
        <option value="">All sources</option>
        <option v-for="t in SOURCE_TYPES.filter((t) => sourceInfo(t).pollable)" :key="t" :value="t">
          {{ sourceInfo(t).label }}
        </option>
      </select>
      <select v-model="tag" class="field w-auto" @change="reload">
        <option value="">All tags</option>
        <option v-for="t in tagsStore.tags" :key="t.id" :value="t.name">#{{ t.name }}</option>
      </select>
      <Toggle v-model="includeNotInUse" label="Include not in use" class="ml-1" @update:model-value="reload" />
    </div>

    <!-- Per-project fetch errors, collapsed and non-fatal -->
    <details v-if="errors.length" class="notice notice-warn mb-5 block">
      <summary class="flex cursor-pointer list-none items-center gap-2">
        <Icon name="alert" />
        {{ errors.length }} project{{ errors.length === 1 ? "" : "s" }} couldn't be fetched
        <Icon name="chevron" class="ml-auto" />
      </summary>
      <ul class="mt-2 flex flex-col gap-1 pl-6 text-xs">
        <li v-for="e in errors" :key="e.project_id">
          <RouterLink :to="`/projects/${e.project_id}`" class="font-medium hover:underline">
            {{ e.project_name }}
          </RouterLink>
          — {{ e.error }}
        </li>
      </ul>
    </details>

    <div v-if="loading && items.length === 0" class="flex flex-col gap-3">
      <div v-for="n in 5" :key="n" class="skeleton h-20" />
    </div>

    <div v-else-if="fatalError" class="notice notice-error">
      <Icon name="alert" class="mt-0.5" /> {{ fatalError }}
    </div>

    <EmptyState
      v-else-if="items.length === 0"
      title="Nothing in the feed yet"
      :hint="
        includeNotInUse
          ? 'Add projects with releases (GitHub, PyPI, crates.io…) and they\'ll show up here.'
          : 'Only projects marked in use appear here. Mark some as in use, or include the rest.'
      "
      icon="feed"
    >
      <button v-if="!includeNotInUse" class="btn btn-outline btn-sm" @click="setIncludeAll">
        Include not in use
      </button>
    </EmptyState>

    <div v-else class="flex flex-col gap-6">
      <section v-for="group in groups" :key="group.label">
        <h2 class="eyebrow sticky top-14 z-10 -mx-1 mb-2 bg-bg/80 px-1 py-2 backdrop-blur-lg">
          {{ group.label }}
        </h2>
        <div class="panel divide-y divide-line overflow-hidden">
          <ReleaseCard
            v-for="(item, i) in group.items"
            :key="`${item.project_id}-${item.release.external_id}-${i}`"
            :item="item"
          />
        </div>
      </section>

      <button v-if="hasMore" class="btn btn-outline mx-auto" :disabled="loading" @click="loadMore">
        <span v-if="loading" class="spinner" />
        Load more
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import EmptyState from "@/components/EmptyState.vue";
import Icon from "@/components/Icon.vue";
import PageHeader from "@/components/PageHeader.vue";
import ReleaseCard from "@/components/ReleaseCard.vue";
import Toggle from "@/components/Toggle.vue";
import { api } from "@/lib/api";
import { dayLabel } from "@/lib/format";
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

// Items arrive sorted newest-first, so consecutive runs share a day label.
const groups = computed(() => {
  const out: { label: string; items: FeedItem[] }[] = [];
  for (const item of items.value) {
    const label = dayLabel(item.release.published_at);
    const last = out[out.length - 1];
    if (last && last.label === label) last.items.push(item);
    else out.push({ label, items: [item] });
  }
  return out;
});

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

function setIncludeAll() {
  includeNotInUse.value = true;
  reload();
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
