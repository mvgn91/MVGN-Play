import { defineConfig } from 'astro/config';
import react from '@astrojs/react';
import tailwindcss from '@tailwindcss/vite';

// Dominio comercial: PENDIENTE_OPERADOR. Mientras no se defina PUBLIC_SITE_URL se
// usa el TLD reservado `.example`, que nunca resuelve, para no inventar un dominio.
const site = process.env.PUBLIC_SITE_URL ?? 'https://mvgnplay.example';

export default defineConfig({
  site,
  integrations: [react()],
  vite: {
    plugins: [tailwindcss()],
  },
  server: {
    port: 4321,
    host: true,
  },
});
