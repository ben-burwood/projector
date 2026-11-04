<template>
  <div class="flex flex-col items-center justify-center gap-3 rounded-2xl border border-dashed border-line px-6 py-16 text-center">
    <div class="flex size-12 items-center justify-center rounded-2xl bg-accent/10 text-accent-text">
      <Icon :name="icon" class="size-6" />
    </div>
    <div class="flex flex-col gap-1">
      <p class="font-semibold">{{ title }}</p>
      <p v-if="hint" class="max-w-sm text-sm text-muted">{{ hint }}</p>
    </div>
    <slot />
  </div>
</template>

<script setup lang="ts">
import Icon from "./Icon.vue";
import type { IconName } from "@/lib/icons";

withDefaults(defineProps<{ title: string; hint?: string; icon?: IconName }>(), {
  icon: "inbox",
});
</script>
