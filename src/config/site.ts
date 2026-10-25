/**
 * Datos del sitio. Fuente única para metadatos, sitemap y textos de contacto.
 *
 * `url`: dominio desplegado en Cloudflare Pages. Se sobrescribe con
 * PUBLIC_SITE_URL si algún día cambia de hosting.
 */
/** `import.meta.env` no existe fuera de Astro (p. ej. en el verificador de Node). */
const env = import.meta.env as Record<string, string | undefined> | undefined;

export const SITE = {
  name: 'MVGN Play',
  parent: 'MVGN Labs',
  /** Marca matriz: relación visible pero secundaria (PRD §2.1). */
  parentUrl: 'https://mvgn.vercel.app',
  locale: 'es-MX',
  url: env?.PUBLIC_SITE_URL ?? 'https://mvgnplay.pages.dev',
  region: 'Zona Metropolitana de Guadalajara, Jalisco',
  description:
    'Configuración profesional de TV Box y ONN Google TV 4K con aplicaciones para ver películas, series y televisión, personalización de interfaz y soporte técnico de 12 meses en la Zona Metropolitana de Guadalajara.',
} as const;

export const isPlaceholderUrl = () => SITE.url.includes('.example');
