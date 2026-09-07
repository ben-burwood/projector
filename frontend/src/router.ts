import { createRouter, createWebHistory } from "vue-router";

const routes = [
  { path: "/", name: "feed", component: () => import("@/views/FeedView.vue") },
  {
    path: "/projects",
    name: "projects",
    component: () => import("@/views/ProjectsView.vue"),
  },
  {
    path: "/projects/:id",
    name: "project",
    component: () => import("@/views/ProjectDetailView.vue"),
  },
  {
    path: "/add",
    name: "add",
    component: () => import("@/views/AddProjectView.vue"),
  },
  { path: "/tags", name: "tags", component: () => import("@/views/TagsView.vue") },
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});
