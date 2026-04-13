import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
	base: '/',
	plugins: [
		tailwindcss(),
		svelte({
			onwarn(warning, defaultHandler) {
				if (warning.code === 'a11y-click-events-have-key-events') return;
				defaultHandler(warning);
			}
		})
	],
	define: {
		__buildDate__: JSON.stringify(new Date().toISOString()),
		'process.browser': true
	},
	build: {
		modulePreload: false,
		sourcemap: true,
		reportCompressedSize: false,
		chunkSizeWarningLimit: 2048
	},
	esbuild: {
		legalComments: 'none'
	},
	server: {
		proxy: {
			'/upload': 'http://localhost:8090',
			'/plans': 'http://localhost:8090'
		}
	}
});
