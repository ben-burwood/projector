<template>
  <span class="chip">
    <span class="text-muted/60">#</span>{{ tag.name }}
    <button
      v-if="removable"
      type="button"
      class="-mr-1 cursor-pointer rounded p-0.5 hover:text-danger"
      aria-label="Remove tag"
      @click.stop="$emit('remove', tag)"
    >
      <Icon name="x" class="size-3" />
    </button>
  </span>
</template>

<script setup lang="ts">
import Icon from "./Icon.vue";
import type { Tag } from "@/lib/types";

defineProps<{ tag: Tag; removable?: boolean }>();
defineEmits<{ remove: [tag: Tag] }>();
</script>
