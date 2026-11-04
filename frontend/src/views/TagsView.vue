<template>
  <div class="mx-auto max-w-xl">
    <PageHeader title="Tags" subtitle="Group projects however you like. Click a tag to see its projects." />

    <form class="mb-5 flex gap-2" @submit.prevent="add">
      <div class="relative flex-1">
        <Icon name="tag" class="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-muted" />
        <input v-model="newName" class="field pl-9" placeholder="New tag name…" />
      </div>
      <button type="submit" class="btn btn-primary" :disabled="!newName.trim()">
        <Icon name="plus" /> Add
      </button>
    </form>

    <div v-if="store.loading && store.tags.length === 0" class="flex flex-col gap-2">
      <div v-for="n in 4" :key="n" class="skeleton h-12" />
    </div>

    <EmptyState
      v-else-if="store.tags.length === 0"
      title="No tags yet"
      hint="Create one above, or add tags while saving a project."
      icon="tag"
    />

    <ul v-else class="panel divide-y divide-line overflow-hidden">
      <li v-for="tag in store.tags" :key="tag.id" class="group flex items-center gap-3 px-4 py-2.5">
        <form v-if="editing === tag.id" class="flex flex-1 items-center gap-2" @submit.prevent="commitRename(tag.id)">
          <input
            v-model="editName"
            class="field h-8"
            autofocus
            @keydown.esc="editing = null"
          />
          <button type="submit" class="btn btn-primary btn-sm btn-icon" aria-label="Save">
            <Icon name="check" />
          </button>
          <button type="button" class="btn btn-ghost btn-sm btn-icon" aria-label="Cancel" @click="editing = null">
            <Icon name="x" />
          </button>
        </form>

        <template v-else>
          <RouterLink
            :to="{ path: '/projects', query: { tag: tag.name } }"
            class="flex min-w-0 flex-1 items-center gap-3"
          >
            <span class="truncate font-medium group-hover:text-accent-text">
              <span class="text-muted">#</span>{{ tag.name }}
            </span>
            <span class="h-1.5 max-w-32 flex-1 overflow-hidden rounded-full bg-surface-2 max-sm:hidden">
              <span
                class="block h-full rounded-full bg-accent/60"
                :style="{ width: `${((tag.project_count ?? 0) / maxCount) * 100}%` }"
              />
            </span>
            <span class="text-xs text-muted tabular-nums">
              {{ tag.project_count ?? 0 }} project{{ tag.project_count === 1 ? "" : "s" }}
            </span>
          </RouterLink>
          <div class="flex gap-0.5 sm:opacity-0 sm:transition-opacity sm:group-hover:opacity-100 sm:focus-within:opacity-100">
            <button class="btn btn-ghost btn-sm btn-icon" aria-label="Rename" title="Rename" @click="startEdit(tag.id, tag.name)">
              <Icon name="pencil" />
            </button>
            <button class="btn btn-danger btn-sm btn-icon" aria-label="Delete" title="Delete" @click="remove(tag.id, tag.name)">
              <Icon name="trash" />
            </button>
          </div>
        </template>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { RouterLink } from "vue-router";
import EmptyState from "@/components/EmptyState.vue";
import Icon from "@/components/Icon.vue";
import PageHeader from "@/components/PageHeader.vue";
import { useTagsStore } from "@/stores/tags";

const store = useTagsStore();
const newName = ref("");
const editing = ref<number | null>(null);
const editName = ref("");

const maxCount = computed(() => Math.max(1, ...store.tags.map((t) => t.project_count ?? 0)));

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

async function remove(id: number, name: string) {
  if (confirm(`Delete #${name}? It will be removed from all projects.`)) {
    await store.remove(id);
  }
}
</script>
