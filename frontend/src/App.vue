<template>
  <div class="bg-base-200 min-h-screen">
    <div class="navbar bg-base-100 shadow-sm">
      <div class="mx-auto flex w-full max-w-5xl items-center gap-2 px-2">
        <RouterLink to="/" class="btn btn-ghost text-xl font-bold">📽️ Projector</RouterLink>
        <div class="flex-1" />
        <nav class="flex items-center gap-1">
          <RouterLink to="/" class="btn btn-ghost btn-sm" exact-active-class="btn-active">
            Feed
          </RouterLink>
          <RouterLink to="/projects" class="btn btn-ghost btn-sm" active-class="btn-active">
            Projects
          </RouterLink>
          <RouterLink to="/tags" class="btn btn-ghost btn-sm" active-class="btn-active">
            Tags
          </RouterLink>
          <RouterLink to="/add" class="btn btn-primary btn-sm">+ Add</RouterLink>
        </nav>
      </div>
    </div>

    <main class="mx-auto w-full max-w-5xl p-4">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView } from "vue-router";
</script>
