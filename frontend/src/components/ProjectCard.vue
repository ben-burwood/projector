<template>
  <div class="card bg-base-100 shadow-md transition-shadow hover:shadow-xl">
    <div class="card-body gap-3 p-4">
      <div class="flex items-start justify-between gap-2">
        <RouterLink
          :to="`/projects/${project.id}`"
          class="link-hover text-lg font-bold leading-tight"
        >
          {{ project.name }}
        </RouterLink>
      </div>

      <div class="flex flex-wrap items-center gap-1">
        <SourceBadge :type="project.source_type" />
        <TagChip v-for="t in project.tags" :key="t.id" :tag="t" />
      </div>

      <div class="card-actions mt-1 items-center justify-between">
        <label class="label cursor-pointer gap-2 py-0">
          <span class="label-text text-xs">In use</span>
          <input
            type="checkbox"
            class="toggle toggle-xs toggle-primary"
            :checked="project.in_use"
            @change="$emit('toggle-inuse', project)"
          />
        </label>
        <a :href="project.url" target="_blank" rel="noopener" class="link link-primary text-xs">
          Open ↗
        </a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterLink } from "vue-router";
import SourceBadge from "./SourceBadge.vue";
import TagChip from "./TagChip.vue";
import type { Project } from "@/lib/types";

defineProps<{ project: Project }>();
defineEmits<{
  "toggle-inuse": [project: Project];
}>();
</script>
