<template>
    <div class="min-h-screen bg-base-200 flex flex-col items-center justify-center p-5">
        <div class="card bg-base-100 shadow-xl w-full max-w-md">
            <div class="card-body items-center text-center gap-4">
                <h1 class="text-4xl font-bold">Hello, World! 👋</h1>
                <p class="text-base-content/70">Projector is up and running.</p>

                <button class="btn btn-primary" :disabled="loading" @click="ping">
                    {{ loading ? "Calling…" : "Call /hello" }}
                </button>

                <div v-if="message" class="alert alert-success">
                    {{ message }}
                </div>
                <div v-if="error" class="alert alert-error">
                    {{ error }}
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { SERVER_URL } from "@/main";

const message = ref("");
const error = ref("");
const loading = ref(false);

async function ping() {
    loading.value = true;
    message.value = "";
    error.value = "";
    try {
        const res = await fetch(`${SERVER_URL}/hello`);
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        const data: { message: string } = await res.json();
        message.value = data.message;
    } catch (err: any) {
        error.value = `Error: ${err.message}`;
    } finally {
        loading.value = false;
    }
}
</script>
