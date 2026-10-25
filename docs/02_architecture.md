# 02 — Arquitectura del Proyecto MVGN Play

> Arquitectura del sitio web comercial de MVGN Play (PRD v2.0).
> Estado: implementado y verificado hasta F5. Régimen de sesión: **SOLO** (Anexo A de `.mvgn/blackboard/JOIN.md`).
> Este documento describe lo que existe en el repositorio, no lo que se planeó.

## Stack real instalado

| Capa | Tecnología | Versión verificada |
|:-----|:-----------|:-------------------|
| Framework | Astro | 7.4.2 (salida estática) |
| Interactividad | React | 19.3.1 con `@astrojs/react` 7.1.1 |
| Estilos | Tailwind CSS | 4.3.3 vía `@tailwindcss/vite` (sin `tailwind.config`) |
| Tipos | TypeScript | 7.0.2 + `@types/node` para scripts |
| Gestor de paquetes | pnpm | 10.33.2 |
| Iconografía / animación | — | Lucide y Motion **no** están instalados; se evalúan si aportan valor (§10.1) |
| Componente React Bits | `CrystalizedBall` (vía `ogl` 1.0.11, WebGL2) | Importado con `npx shadcn@latest add @react-bits/CrystalizedBall-TS-TW`; requiere `components.json` |

## Árbol de `src/`

```text
src/
├── components/
│   ├── hero/
│   │   ├── Hero.astro              # §9.2 composición editorial, CTAs y datos
│   │   ├── BallHero.tsx            # adaptador de tema/movimiento para la bola
│   │   └── CrystalizedBall.tsx     # React Bits (ogl) — componente importado
│   ├── showcase/
│   │   ├── ShowcaseSection.astro   # §9.3 contenedor (id="servicio")
│   │   └── ExperienceShowcase.tsx  # isla React: tablist ARIA de 3 estados
│   ├── commercial/
│   │   ├── DeviceSection.astro     # §9.4
│   │   ├── HowItWorks.astro        # §9.5
│   │   ├── SupportSection.astro    # §9.7
│   │   ├── CompatibilitySection.astro # §9.8
│   │   ├── FaqSection.astro        # §9.9 acordeón <details> nativo
│   │   └── FinalCta.astro          # §9.10
│   ├── pricing/
│   │   ├── PricingSection.astro    # §9.6 contenedor (id="precios")
│   │   └── PriceSelector.tsx       # isla React: cálculo + CTA contextual
│   └── ui/
│       ├── WhatsAppCta.astro       # CTA con respaldo honesto si falta el número
│       └── ThemeToggle.astro       # conmutador claro/oscuro persistente
├── config/
│   ├── commerce.ts                 # FUENTE ÚNICA de precios, vigencia y mensajes §12
│   ├── pricing.ts                  # alias heredados derivados (sin importes literales)
│   └── site.ts                     # nombre, región, URL y descripción
├── layouts/
│   ├── BaseLayout.astro            # head, canonical, Open Graph, tema inicial, skip-link
│   ├── PageLayout.astro            # páginas documentales
│   ├── SiteNav.astro               # §9.1 cápsula flotante + menú a pantalla completa
│   └── SiteFooter.astro            # §9.11 wordmark + columnas + barra legal
├── pages/
│   ├── index.astro                 # landing, orden §8.1
│   ├── soporte.astro               # §5
│   ├── condiciones.astro           # §4 / §12
│   ├── privacidad.astro            # borrador
│   ├── 404.astro
│   └── sitemap.xml.ts              # endpoint de sitemap
├── styles/
│   ├── tokens.css                  # paleta §7.2, tipografía §7.3, capa base y .doc
│   └── global.css                  # entrada
└── env.d.ts
```

Fuera de `src/`: `public/` (favicon, `og.png`, `robots.txt`), `scripts/verify-pricing.ts`
(verificador del §12.1) y `.env.example` (variables documentadas).

## Decisiones clave

| Decisión | Implementación real |
|:---------|:--------------------|
| Astro para estructura y contenido | Todo el contenido se renderiza en HTML; no hay isla en el hero |
| React solo donde hay estado real | Dos islas: demo interactiva y selector de precios, ambas con `client:visible` |
| Tailwind con tokens | `@theme` en `tokens.css`; ningún hex suelto en componentes |
| Precios desde una sola fuente | `src/config/commerce.ts` con `quote()`; alias derivados en `pricing.ts`; verificado por `pnpm verify` |
| WhatsApp como único CTA comercial | `whatsappLink()` devuelve `null` sin número: se muestra aviso, nunca un enlace inventado |
| Contenido funciona sin JS | Nav, FAQ (`<details>`), precios visibles y CTA de respaldo no dependen de JavaScript |
| Movimiento respetuoso | `prefers-reduced-motion` global + tilt 3D solo con `pointer: fine` |
| Tema claro/oscuro | Variables semánticas en `:root` y `[data-theme='light']`; `@theme inline` mapea a `var()`; `data-theme` en `<html>`, persistido en `localStorage` y aplicado antes del primer pintado |
| Navegación | Cápsula flotante con blur; en móvil menú a pantalla completa con foco atrapado, ESC y liberación del scroll |
| Sin objetos simulados | El hero es composición editorial con la bola crystalizada (React Bits, WebGL vía `ogl`); la sección del equipo compara los dos caminos en lugar de dibujar hardware |
| WebGL deliberado | `CrystalizedBall` usa WebGL2; se carga con `client:idle`, se pausa con `prefers-reduced-motion`, baja a 6000 partículas en móvil y degrada a vacío si no hay WebGL2 |

## Variables de entorno

| Variable | Efecto si falta |
|:---------|:----------------|
| `PUBLIC_SITE_URL` | `site` cae a `https://mvgnplay.example` (TLD reservado): canonical, sitemap y OG quedan en placeholder |
| `PUBLIC_WHATSAPP_NUMBER` | Los CTA muestran el aviso de canal pendiente en lugar de un enlace |

## Verificación reproducible

| Comprobante | Comando | Resultado esperado |
|:------------|:--------|:-------------------|
| Build | `pnpm build` | 5 rutas + sitemap, sin errores |
| Tipos | `pnpm typecheck` | exit 0 |
| Precios §12.1 | `pnpm verify` | los 4 totales + mensajes + alias heredados |
| Rutas | `curl` a `/`, `/soporte`, `/condiciones`, `/privacidad`, `/sitemap.xml` | 200; `/no-existe` → 404 personalizada |

## Pendiente (no arquitectónico, sí de publicación)

- Lighthouse / Chromium y prueba en navegador móvil: no medidos (§14, criterios 12–14).
- Número comercial, dominio y datos del responsable de privacidad.
- Revisión legal de `/condiciones` y `/privacidad` antes de publicar.

---

*Documento corregido por ATHN-01 el 2026-10-25 para reflejar el repositorio real; la versión anterior describía carpetas que no existen.*
