import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
export default defineConfig({
    plugins: [vue(), ],
    define: {
        __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: 'false',

    },
    resolve: {
        alias: {
            '@': fileURLToPath(new URL('./src', import.meta.url)),

        },
    },
})