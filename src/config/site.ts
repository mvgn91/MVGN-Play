/**
 * Datos del sitio. Fuente única para metadatos, sitemap y textos de contacto.
 *
 * `url`: PENDIENTE_OPERADOR. El dominio comercial no lo entrega el PRD, así que
 * se usa el TLD reservado `.example` (nunca resuelve) hasta que el operador lo
 * defina en PUBLIC_SITE_URL. Ningún componente debe escribir un dominio a mano.
 */
/** `import.meta.env` no existe fuera de Astro (p. ej. en el verificador de Node). */
const env = import.meta.env as Record<string, string | undefined> | undefined;

export const SITE = {
  name: 'MVGN Play',
  parent: 'MVGN Labs',
  /** Marca matriz: relación visible pero secundaria (PRD §2.1). */
  parentUrl: 'https://mvgn.vercel.app',
  locale: 'es-MX',
  url: env?.PUBLIC_SITE_URL ?? 'https://mvgnplay.example',
  region: 'Zona Metropolitana de Guadalajara, Jalisco',
  description:
    'Configuración profesional de TV Box y ONN Google TV 4K con aplicaciones para ver películas, series y televisión, personalización de interfaz y soporte técnico de 12 meses en la Zona Metropolitana de Guadalajara.',
} as const;

export const isPlaceholderUrl = () => SITE.url.includes('.example');
