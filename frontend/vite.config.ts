import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// https://vite.dev/config/
export default defineConfig({
	plugins: [react()],
	server: {
		host: true, // Listen on all network interfaces
		port: 5173, // Default Vite port
		strictPort: true, // Exit if the port is already in use
	},
});
