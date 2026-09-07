<template>
  <div class="card bg-base-100 shadow-sm">
    <div class="card-body gap-2 p-4">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div class="flex items-center gap-2">
          <RouterLink
            :to="`/projects/${item.project_id}`"
            class="link-hover font-semibold"
          >
            {{ item.project_name }}
          </RouterLink>
          <SourceBadge :type="item.source_type" />
        </div>
        <span class="text-base-content/50 text-xs">{{ published }}</span>
      </div>

      <div class="flex items-baseline gap-2">
        <span v-if="item.release.version" class="badge badge-primary badge-sm font-mono">
          {{ item.release.version }}
        </span>
        <span v-if="item.release.title" class="text-sm font-medium">{{ item.release.title }}</span>
      </div>

      <p v-if="excerpt" class="text-base-content/70 whitespace-pre-line text-sm">
        {{ excerpt }}
      </p>

      <div class="card-actions justify-end">
        <a
          v-if="item.release.url"
          :href="item.release.url"
          target="_blank"
          rel="noopener"
          class="link link-primary text-xs"
        >
          View release ↗
        </a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RouterLink } from "vue-router";
import SourceBadge from "./SourceBadge.vue";
import type { FeedItem } from "@/lib/types";

const props = defineProps<{ item: FeedItem }>();

const published = computed(() => {
  const p = props.item.release.published_at;
  if (!p) return "";
  const d = new Date(p);
  if (isNaN(d.getTime())) return "";
  return d.toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
});

const excerpt = computed(() => {
  const body = props.item.release.body ?? "";
  return body.length > 280 ? body.slice(0, 280) + "…" : body;
});
</script>
