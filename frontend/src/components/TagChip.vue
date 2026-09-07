<template>
  <span class="badge badge-outline badge-sm gap-1">
    {{ tag.name }}
    <button
      v-if="removable"
      type="button"
      class="hover:text-error"
      aria-label="Remove tag"
      @click.stop="$emit('remove', tag)"
    >
      ✕
    </button>
  </span>
</template>

<script setup lang="ts">
import type { Tag } from "@/lib/types";

defineProps<{ tag: Tag; removable?: boolean }>();
defineEmits<{ remove: [tag: Tag] }>();
</script>
