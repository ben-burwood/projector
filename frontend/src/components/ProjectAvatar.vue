<template>
  <img
    v-if="avatarUrl && !failed"
    :src="avatarUrl"
    alt=""
    loading="lazy"
    class="shrink-0 rounded-lg border border-line bg-surface-2 object-cover"
    :class="sizeClass"
    @error="failed = true"
  />
  <span
    v-else
    class="inline-flex shrink-0 items-center justify-center rounded-lg font-semibold uppercase"
    :class="sizeClass"
    :style="{
      color: info.color,
      backgroundColor: `color-mix(in srgb, ${info.color} 16%, transparent)`,
    }"
    aria-hidden="true"
  >
    {{ initial }}
  </span>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { sourceInfo } from "@/lib/sources";
import type { SourceType } from "@/lib/types";

const props = withDefaults(
  defineProps<{ name: string; type: SourceType; identifier?: string; size?: "sm" | "md" | "lg" }>(),
  { size: "md" },
);

const failed = ref(false);
const info = computed(() => sourceInfo(props.type));
const initial = computed(() => props.name.trim().charAt(0) || "?");

// GitHub serves owner avatars at a predictable URL; everything else gets a
// tinted initial.
const avatarUrl = computed(() => {
  if (props.type !== "github" || !props.identifier) return "";
  const owner = props.identifier.split("/")[0];
  return owner ? `https://github.com/${owner}.png?size=96` : "";
});

const sizeClass = computed(
  () => ({ sm: "size-8 text-sm", md: "size-10 text-base", lg: "size-14 text-2xl" })[props.size],
);
</script>
