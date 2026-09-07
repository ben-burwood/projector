<template>
  <span class="badge badge-sm" :class="info.badge">{{ info.label }}</span>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { sourceInfo } from "@/lib/sources";
import type { SourceType } from "@/lib/types";

const props = defineProps<{ type: SourceType }>();
const info = computed(() => sourceInfo(props.type));
</script>
