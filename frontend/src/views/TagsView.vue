<template>
  <div class="mx-auto flex max-w-lg flex-col gap-4">
    <h1 class="text-2xl font-bold">Tags</h1>

    <div class="join w-full">
      <input
        v-model="newName"
        class="input input-bordered join-item w-full"
        placeholder="New tag name…"
        @keydown.enter.prevent="add"
      />
      <button class="btn btn-primary join-item" :disabled="!newName.trim()" @click="add">Add</button>
    </div>

    <div v-if="store.loading" class="flex justify-center py-8">
      <span class="loading loading-spinner" />
    </div>

    <EmptyState v-else-if="store.tags.length === 0" title="No tags yet" icon="🏷️" />

    <ul v-else class="menu bg-base-100 rounded-box w-full shadow">
      <li v-for="tag in store.tags" :key="tag.id">
        <div class="flex items-center justify-between">
          <template v-if="editing === tag.id">
            <input
              v-model="editName"
              class="input input-bordered input-sm flex-1"
              @keydown.enter.prevent="commitRename(tag.id)"
              @keydown.esc="editing = null"
            />
            <div class="flex gap-1">
              <button class="btn btn-xs btn-primary" @click="commitRename(tag.id)">Save</button>
              <button class="btn btn-xs" @click="editing = null">Cancel</button>
            </div>
          </template>
          <template v-else>
            <span>{{ tag.name }}<span class="text-base-content/50 ml-2 text-xs">({{ tag.project_count ?? 0 }})</span></span>
            <div class="flex gap-1">
              <button class="btn btn-ghost btn-xs" @click="startEdit(tag.id, tag.name)">Rename</button>
              <button class="btn btn-ghost btn-xs text-error" @click="remove(tag.id)">Delete</button>
            </div>
          </template>
        </div>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import EmptyState from "@/components/EmptyState.vue";
import { useTagsStore } from "@/stores/tags";

const store = useTagsStore();
const newName = ref("");
const editing = ref<number | null>(null);
const editName = ref("");

onMounted(() => store.load(true));

async function add() {
  const name = newName.value.trim();
  if (!name) return;
  await store.create(name);
  newName.value = "";
}

function startEdit(id: number, name: string) {
  editing.value = id;
  editName.value = name;
}

async function commitRename(id: number) {
  const name = editName.value.trim();
  if (name) await store.rename(id, name);
  editing.value = null;
}

async function remove(id: number) {
  if (confirm("Delete this tag? It will be removed from all projects.")) {
    await store.remove(id);
  }
}
</script>
