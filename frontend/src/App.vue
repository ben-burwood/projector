<template>
  <div class="min-h-screen">
    <header class="sticky top-0 z-30 border-b border-line bg-bg/80 backdrop-blur-lg">
      <div class="mx-auto flex h-14 w-full max-w-5xl items-center gap-3 px-4">
        <RouterLink to="/" class="flex items-center gap-2.5 font-semibold tracking-tight">
          <img src="/favicon.svg" alt="" class="size-7" />
          <span class="hidden text-[0.95rem] sm:inline">Projector</span>
        </RouterLink>

        <nav class="ml-2 flex items-center gap-0.5 sm:ml-4">
          <RouterLink
            v-for="item in NAV"
            :key="item.to"
            :to="item.to"
            class="relative flex h-9 items-center gap-1.5 rounded-lg px-2.5 text-sm text-muted transition-colors hover:text-ink"
            :class="isActive(item.to) && 'text-ink'"
          >
            <Icon :name="item.icon" class="hidden sm:block" />
            {{ item.label }}
            <span
              v-if="isActive(item.to)"
              class="absolute inset-x-2.5 -bottom-[11px] h-0.5 rounded-full bg-accent"
            />
          </RouterLink>
        </nav>

        <div class="flex-1" />
        <RouterLink to="/add" class="btn btn-primary btn-sm">
          <Icon name="plus" />
          <span class="hidden sm:inline">Add</span>
        </RouterLink>
      </div>
    </header>

    <main class="mx-auto w-full max-w-5xl px-4 py-6 sm:py-8">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView, useRoute } from "vue-router";
import Icon from "@/components/Icon.vue";
import type { IconName } from "@/lib/icons";

const NAV: { to: string; label: string; icon: IconName }[] = [
  { to: "/", label: "Feed", icon: "feed" },
  { to: "/projects", label: "Projects", icon: "folder" },
  { to: "/tags", label: "Tags", icon: "tag" },
];

const route = useRoute();

function isActive(to: string): boolean {
  return to === "/" ? route.path === "/" : route.path.startsWith(to);
}
</script>
