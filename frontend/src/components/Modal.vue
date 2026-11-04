<template>
  <dialog
    ref="dialog"
    class="m-auto w-[calc(100%-2rem)] max-w-md rounded-2xl border border-line bg-surface p-0 text-ink shadow-2xl backdrop:bg-black/50 backdrop:backdrop-blur-sm"
    @close="$emit('update:open', false)"
    @click.self="$emit('update:open', false)"
  >
    <div class="flex flex-col gap-4 p-5">
      <div class="flex items-center justify-between">
        <h3 class="text-base font-semibold">{{ title }}</h3>
        <button class="btn btn-ghost btn-sm btn-icon" aria-label="Close" @click="$emit('update:open', false)">
          <Icon name="x" />
        </button>
      </div>
      <slot />
      <div v-if="$slots.actions" class="flex justify-end gap-2 pt-1">
        <slot name="actions" />
      </div>
    </div>
  </dialog>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import Icon from "./Icon.vue";

const props = defineProps<{ open: boolean; title: string }>();
defineEmits<{ "update:open": [open: boolean] }>();

const dialog = ref<HTMLDialogElement>();

watch(
  () => props.open,
  (open) => {
    if (open) dialog.value?.showModal();
    else dialog.value?.close();
  },
);
</script>
