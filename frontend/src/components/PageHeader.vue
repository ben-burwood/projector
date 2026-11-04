<template>
  <div class="mb-6 flex flex-wrap items-end justify-between gap-3">
    <div class="flex flex-col gap-1">
      <h1 class="text-2xl font-semibold tracking-tight">{{ title }}</h1>
      <p v-if="subtitle || $slots.subtitle" class="text-sm text-muted">
        <slot name="subtitle">{{ subtitle }}</slot>
      </p>
    </div>
    <div v-if="$slots.default" class="flex items-center gap-2">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ title: string; subtitle?: string }>();
</script>
