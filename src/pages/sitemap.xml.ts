import type { APIRoute } from 'astro';
import { SITE, isPlaceholderUrl } from '@/config/site';

/** Rutas indexables. Las 404 no entran al sitemap. */
const ROUTES: Array<{ path: string; priority: string; changefreq: string }> = [
  { path: '/', priority: '1.0', changefreq: 'monthly' },
  { path: '/soporte', priority: '0.7', changefreq: 'yearly' },
  { path: '/condiciones', priority: '0.6', changefreq: 'yearly' },
  { path: '/privacidad', priority: '0.4', changefreq: 'yearly' },
];

export const GET: APIRoute = () => {
  const base = SITE.url.replace(/\/$/, '');
  const lastmod = new Date().toISOString().slice(0, 10);

  const urls = ROUTES.map(
    (route) => `  <url>
    <loc>${base}${route.path === '/' ? '/' : route.path}</loc>
    <lastmod>${lastmod}</lastmod>
    <changefreq>${route.changefreq}</changefreq>
    <priority>${route.priority}</priority>
  </url>`,
  ).join('\n');

  const note = isPlaceholderUrl()
    ? '<!-- SITE_URL PENDIENTE_OPERADOR: define PUBLIC_SITE_URL antes de publicar. -->\n'
    : '';

  return new Response(
    `<?xml version="1.0" encoding="UTF-8"?>\n${note}<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n${urls}\n</urlset>\n`,
    {
      headers: {
        'Content-Type': 'application/xml; charset=utf-8',
        'Cache-Control': 'public, max-age=3600',
      },
    },
  );
};
