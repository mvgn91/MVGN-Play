import { defineConfig } from 'astro/config';
import react from '@astrojs/react';
import tailwindcss from '@tailwindcss/vite';

// Dominio desplegado. PUBLIC_SITE_URL lo sobrescribe si cambia de hosting;
// sin variable, el fallback es el dominio real para que canonical, sitemap y
// Open Graph nunca queden en placeholder.
const site = process.env.PUBLIC_SITE_URL ?? 'https://mvgnplay.pages.dev';

export default defineConfig({
  site,
  // `file` deja las rutas sin barra final (/soporte en vez de /soporte/):
  // Cloudflare Pages sirve esos HTML sin el redirect 308 que sí haría con
  // el formato `directory`, y así canonical, sitemap y enlaces internos
  // coinciden sin redirecciones.
  build: {
    format: 'file',
  },
  integrations: [react()],
  vite: {
    plugins: [tailwindcss()],
  },
  server: {
    port: 4321,
    host: true,
  },
});
