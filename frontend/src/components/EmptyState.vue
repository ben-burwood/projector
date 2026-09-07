<template>
  <div class="flex flex-col items-center justify-center gap-2 py-16 text-center">
    <div class="text-4xl opacity-40">{{ icon }}</div>
    <p class="text-lg font-semibold">{{ title }}</p>
    <p v-if="hint" class="text-base-content/60 max-w-sm text-sm">{{ hint }}</p>
    <slot />
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ title: string; hint?: string; icon?: string }>(), {
  icon: "📭",
});
</script>
