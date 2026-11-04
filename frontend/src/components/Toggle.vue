<template>
  <button
    type="button"
    role="switch"
    :aria-checked="modelValue"
    :aria-label="label"
    class="group inline-flex cursor-pointer items-center gap-2 text-xs text-muted select-none"
    @click.stop.prevent="$emit('update:modelValue', !modelValue)"
  >
    <span
      class="relative inline-flex h-5 w-9 shrink-0 items-center rounded-full border transition-colors"
      :class="modelValue ? 'border-accent bg-accent' : 'border-line bg-surface-2 group-hover:border-muted/50'"
    >
      <span
        class="size-3.5 rounded-full shadow-sm transition-transform"
        :class="modelValue ? 'translate-x-[1.05rem] bg-on-accent' : 'translate-x-0.5 bg-muted/70'"
      />
    </span>
    <span v-if="label && showLabel">{{ label }}</span>
  </button>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ modelValue: boolean; label?: string; showLabel?: boolean }>(), {
  showLabel: true,
});
defineEmits<{ "update:modelValue": [value: boolean] }>();
</script>
