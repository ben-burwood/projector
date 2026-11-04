<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
    width="16"
    height="16"
    class="shrink-0"
    aria-hidden="true"
    v-html="ICONS[name]"
  />
</template>

<script setup lang="ts">
import { ICONS, type IconName } from "@/lib/icons";

defineProps<{ name: IconName }>();
</script>
