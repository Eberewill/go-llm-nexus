import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, process.cwd(), '')
	const trimmedBase = (env.VITE_API_BASE || '').replace(/\/$/, '')
	const proxyTarget = trimmedBase || env.VITE_DEV_SERVER_API || 'http://localhost:8080'

	return {
		plugins: [react()],
		server: {
			port: 3001,
			proxy: trimmedBase
				? undefined
				: {
					'/api': {
						target: proxyTarget,
						changeOrigin: true,
					},
				},
		},
	}
})
