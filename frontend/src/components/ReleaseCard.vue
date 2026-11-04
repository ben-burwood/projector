<template>
  <article class="group flex gap-3 px-4 py-3.5 transition-colors hover:bg-surface-2/50 sm:gap-4">
    <RouterLink :to="`/projects/${item.project_id}`" class="pt-0.5">
      <ProjectAvatar
        :name="item.project_name"
        :type="item.source_type"
        :identifier="item.source_identifier"
        size="sm"
      />
    </RouterLink>

    <div class="flex min-w-0 flex-1 flex-col gap-1">
      <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
        <RouterLink :to="`/projects/${item.project_id}`" class="font-medium hover:text-accent-text">
          {{ item.project_name }}
        </RouterLink>
        <span v-if="item.release.version" class="version">{{ item.release.version }}</span>
        <span class="ml-auto text-xs text-muted tabular-nums" :title="fmtDate(item.release.published_at)">
          {{ relativeTime(item.release.published_at) }}
        </span>
      </div>

      <p v-if="showTitle" class="text-sm text-ink/90">{{ item.release.title }}</p>

      <p
        v-if="excerpt"
        class="text-sm whitespace-pre-line text-muted"
        :class="expanded ? '' : 'line-clamp-2'"
      >
        {{ excerpt }}
      </p>

      <div class="flex items-center gap-3 text-xs">
        <button
          v-if="excerpt.length > 160"
          type="button"
          class="cursor-pointer text-muted hover:text-ink"
          @click="expanded = !expanded"
        >
          {{ expanded ? "Show less" : "Show more" }}
        </button>
        <a
          v-if="item.release.url"
          :href="item.release.url"
          target="_blank"
          rel="noopener"
          class="inline-flex items-center gap-1 text-muted opacity-0 transition-opacity group-hover:opacity-100 hover:text-accent-text focus:opacity-100 max-sm:opacity-100"
        >
          View release <Icon name="external" class="size-3" />
        </a>
      </div>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { RouterLink } from "vue-router";
import Icon from "./Icon.vue";
import ProjectAvatar from "./ProjectAvatar.vue";
import { fmtDate, plainText, relativeTime } from "@/lib/format";
import type { FeedItem } from "@/lib/types";

const props = defineProps<{ item: FeedItem }>();

const expanded = ref(false);

// Hide the title when it just repeats the version and/or project name
// ("v1.2.0", "beszel v1.2.0") — common for tag-only and auto-named releases.
const showTitle = computed(() => {
  const { title, version } = props.item.release;
  if (!title) return false;
  const norm = (s: string) => s.toLowerCase().replace(/^v/, "").trim();
  const rest = norm(title.replace(version, "").replace(props.item.project_name, ""));
  return rest !== "" && norm(title) !== norm(version);
});

const excerpt = computed(() => plainText(props.item.release.body));
</script>
