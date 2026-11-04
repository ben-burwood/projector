<template>
  <div class="flex flex-wrap items-center gap-1.5">
    <button
      v-for="tag in tagsStore.tags"
      :key="tag.id"
      type="button"
      class="chip cursor-pointer transition-colors"
      :class="
        modelValue.includes(tag.id)
          ? 'border-accent bg-accent/15 text-accent-text'
          : 'hover:border-muted/50 hover:text-ink'
      "
      :aria-pressed="modelValue.includes(tag.id)"
      @click="toggle(tag.id)"
    >
      <Icon v-if="modelValue.includes(tag.id)" name="check" class="size-3" />
      {{ tag.name }}
    </button>

    <form class="flex items-center" @submit.prevent="addNew">
      <input
        v-model="newTag"
        class="h-6 w-24 rounded-md border border-dashed border-line bg-transparent px-2 text-xs outline-none placeholder:text-muted/70 focus:border-accent focus:w-32 transition-[width]"
        placeholder="+ new tag"
      />
    </form>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import Icon from "./Icon.vue";
import { useTagsStore } from "@/stores/tags";

const props = defineProps<{ modelValue: number[] }>();
const emit = defineEmits<{ "update:modelValue": [ids: number[]] }>();

const tagsStore = useTagsStore();
const newTag = ref("");

onMounted(() => tagsStore.load());

function toggle(id: number) {
  const set = new Set(props.modelValue);
  set.has(id) ? set.delete(id) : set.add(id);
  emit("update:modelValue", [...set]);
}

async function addNew() {
  const name = newTag.value.trim();
  if (!name) return;
  const tag = await tagsStore.create(name);
  newTag.value = "";
  if (!props.modelValue.includes(tag.id)) {
    emit("update:modelValue", [...props.modelValue, tag.id]);
  }
}
</script>
