/**
 * Alias heredados de la configuración comercial.
 *
 * FUENTE ÚNICA: `src/config/commerce.ts` (PRD §12.1). Este archivo existe solo
 * para no romper importadores antiguos y **no declara ni un importe**: todos los
 * valores se derivan del cálculo canónico. Si necesitas precios nuevos, agrega
 * la lógica en `commerce.ts`, nunca aquí.
 *
 * @deprecated Usa `@/config/commerce`.
 */
import { PRICING, SUPPORT_MONTHS, quote } from './commerce.ts';

/** Nombres anteriores de los importes, resueltos contra la fuente única. */
export const LEGACY_PRICING = {
  ONN_CONFIGURED: PRICING.equipmentIncluded,
  CLIENT_EQUIPMENT: PRICING.customerEquipment,
  INSTALLATION_PRESENCE: PRICING.onsiteInstallation,
  SUPPORT_DURATION_MONTHS: SUPPORT_MONTHS,
} as const;

export type PricingKey = keyof typeof LEGACY_PRICING;

/** Totales anteriores, calculados — nunca escritos a mano. */
export const PRICING_TOTALS = {
  onnNoInstall: quote({ equipment: 'included', installation: 'none' }).total,
  onnInstall: quote({ equipment: 'included', installation: 'onsite' }).total,
  clientNoInstall: quote({ equipment: 'customer', installation: 'none' }).total,
  clientInstall: quote({ equipment: 'customer', installation: 'onsite' }).total,
} as const;

export { PRICING, SUPPORT_MONTHS } from './commerce.ts';
export { formatMXN, quote, buildMessage, whatsappLink } from './commerce.ts';
