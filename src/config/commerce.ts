/**
 * Configuración comercial centralizada — PRD §12.1 y §12.2.
 *
 * Ningún importe ni mensaje debe duplicarse en componentes: todo se calcula o
 * se compone desde aquí. Si cambia un precio, cambia en un solo lugar.
 */

export const PRICING = {
  /** Modalidad A — equipo ONN 4K configurado. */
  equipmentIncluded: 2000,
  /** Modalidad B — configuración y soporte para equipo del cliente. */
  customerEquipment: 1000,
  /** Modalidad C — cargo adicional por instalación presencial. */
  onsiteInstallation: 500,
} as const;

/** Vigencia del soporte técnico incluido en las modalidades A y B (PRD §5.1). */
export const SUPPORT_MONTHS = 12;

export const CURRENCY = 'MXN';

export type Equipment = 'included' | 'customer';
export type Installation = 'none' | 'onsite';

export interface QuoteInput {
  equipment: Equipment;
  installation: Installation;
}

export interface QuoteLine {
  label: string;
  amount: number;
}

export interface Quote {
  lines: QuoteLine[];
  total: number;
  currency: typeof CURRENCY;
}

/** Formato monetario único del proyecto: `$2,000 MXN`. */
export function formatMXN(amount: number): string {
  const value = new Intl.NumberFormat('es-MX', {
    style: 'currency',
    currency: CURRENCY,
    maximumFractionDigits: 0,
  }).format(amount);

  return `${value} ${CURRENCY}`;
}

/**
 * Calcula el total a partir de la modalidad y la necesidad de instalación.
 * Casos obligatorios del PRD §12.1: 2000 / 2500 / 1000 / 1500.
 * El traslado fuera de la ZMG no se suma: requiere cotización (PRD §4.3).
 */
export function quote({ equipment, installation }: QuoteInput): Quote {
  const base = equipment === 'included' ? PRICING.equipmentIncluded : PRICING.customerEquipment;
  const label =
    equipment === 'included' ? 'Equipo ONN 4K configurado' : 'Configuración de tu propio equipo';

  const lines: QuoteLine[] = [{ label, amount: base }];

  if (installation === 'onsite') {
    lines.push({ label: 'Instalación presencial', amount: PRICING.onsiteInstallation });
  }

  return {
    lines,
    total: lines.reduce((sum, line) => sum + line.amount, 0),
    currency: CURRENCY,
  };
}

/**
 * Número comercial entregado por el operador: +52 1 33 2262 1939.
 * Se guarda en formato internacional sin signos, como lo espera wa.me.
 * PUBLIC_WHATSAPP_NUMBER lo sobrescribe si el número cambia.
 * `import.meta.env` no existe fuera de Astro (verificador de Node).
 */
const env = import.meta.env as Record<string, string | undefined> | undefined;

export const WHATSAPP_NUMBER = env?.PUBLIC_WHATSAPP_NUMBER ?? '5213322621939';

export const isWhatsAppConfigured = () => /^\d{10,15}$/.test(WHATSAPP_NUMBER);

/** Número en formato legible para mostrarlo en pantalla. */
export function formatPhone(raw: string = WHATSAPP_NUMBER): string {
  if (!/^\d{12,13}$/.test(raw)) return raw;
  const local = raw.slice(-10);
  return `+52 ${local.slice(0, 2)} ${local.slice(2, 6)} ${local.slice(6)}`;
}

/**
 * Mensajes prellenados — base textual del PRD §12.2. Los importes se toman de
 * PRICING para que el mensaje nunca contradiga al calculado.
 */
export function buildMessage({ equipment, installation }: QuoteInput): string {
  const { total } = quote({ equipment, installation });

  const base =
    equipment === 'included'
      ? `Hola, me interesa la ONN 4K configurada de ${formatMXN(PRICING.equipmentIncluded)}. ¿Podrías confirmarme disponibilidad y condiciones de entrega?`
      : `Hola, me interesa la configuración y soporte por ${SUPPORT_MONTHS} meses para mi propia TV Box por ${formatMXN(PRICING.customerEquipment)}. ¿Podemos verificar la compatibilidad?`;

  if (installation === 'none') return base;

  return `${base} También me interesa la instalación presencial de ${formatMXN(PRICING.onsiteInstallation)} (total ${formatMXN(total)}): ¿podrías confirmar la cobertura para mi domicilio?`;
}

/** Enlace de WhatsApp contextual. `null` mientras el número no esté configurado. */
export function whatsappLink(input: QuoteInput = { equipment: 'included', installation: 'none' }): string | null {
  if (!isWhatsAppConfigured()) return null;

  const text = encodeURIComponent(buildMessage(input));
  return `https://wa.me/${WHATSAPP_NUMBER}?text=${text}`;
}

/** Destino de respaldo mientras no hay número comercial: nunca un enlace roto. */
export const CONTACT_FALLBACK_HREF = '#consulta';
