<template>
  <div class="group flex items-center gap-3 px-4 py-3 transition-colors hover:bg-surface-2/50">
    <ProjectAvatar :name="project.name" :type="project.source_type" :identifier="project.source_identifier" />

    <RouterLink :to="`/projects/${project.id}`" class="flex min-w-0 flex-1 flex-col gap-1">
      <div class="flex min-w-0 items-baseline gap-2">
        <span class="truncate font-medium group-hover:text-accent-text">{{ project.name }}</span>
        <span
          v-if="project.source_identifier && project.source_identifier !== project.name"
          class="hidden truncate font-mono text-xs text-muted md:inline"
        >
          {{ project.source_identifier }}
        </span>
      </div>
      <div class="flex flex-wrap items-center gap-x-3 gap-y-1">
        <SourceBadge :type="project.source_type" />
        <span v-for="t in project.tags" :key="t.id" class="text-xs text-muted">#{{ t.name }}</span>
      </div>
    </RouterLink>

    <Toggle
      :model-value="project.in_use"
      label="In use"
      :show-label="false"
      title="In use"
      @update:model-value="$emit('toggle-inuse', project)"
    />
    <a
      :href="project.url"
      target="_blank"
      rel="noopener"
      class="btn btn-ghost btn-sm btn-icon"
      :aria-label="`Open ${project.name}`"
    >
      <Icon name="external" />
    </a>
  </div>
</template>

<script setup lang="ts">
import { RouterLink } from "vue-router";
import Icon from "./Icon.vue";
import ProjectAvatar from "./ProjectAvatar.vue";
import SourceBadge from "./SourceBadge.vue";
import Toggle from "./Toggle.vue";
import type { Project } from "@/lib/types";

defineProps<{ project: Project }>();
defineEmits<{
  "toggle-inuse": [project: Project];
}>();
</script>
