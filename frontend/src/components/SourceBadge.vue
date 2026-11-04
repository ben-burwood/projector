<template>
  <span class="inline-flex items-center gap-1.5 text-xs text-muted">
    <span class="size-2 rounded-full" :style="{ backgroundColor: info.color }" />
    {{ info.label }}
  </span>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { sourceInfo } from "@/lib/sources";
import type { SourceType } from "@/lib/types";

const props = defineProps<{ type: SourceType }>();
const info = computed(() => sourceInfo(props.type));
</script>
