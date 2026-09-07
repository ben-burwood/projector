<template>
  <div class="flex flex-wrap items-center gap-2">
    <label
      v-for="tag in tagsStore.tags"
      :key="tag.id"
      class="cursor-pointer"
    >
      <input
        type="checkbox"
        class="peer sr-only"
        :checked="modelValue.includes(tag.id)"
        @change="toggle(tag.id)"
      />
      <span
        class="badge badge-outline peer-checked:badge-primary peer-checked:badge-outline-0"
        :class="modelValue.includes(tag.id) ? 'badge-primary' : ''"
      >
        {{ tag.name }}
      </span>
    </label>

    <div class="join">
      <input
        v-model="newTag"
        class="input input-xs input-bordered join-item w-28"
        placeholder="new tag…"
        @keydown.enter.prevent="addNew"
      />
      <button type="button" class="btn btn-xs join-item" @click="addNew">+</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
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
