/**
 * Verificador de la configuración comercial — PRD §12.1.
 * Ejecuta: node scripts/verify-pricing.ts
 * Falla (exit 1) si un caso obligatorio no coincide con el PRD.
 */
import {
  PRICING,
  SUPPORT_MONTHS,
  WHATSAPP_NUMBER,
  buildMessage,
  formatMXN,
  formatPhone,
  isWhatsAppConfigured,
  quote,
  whatsappLink,
  type Equipment,
  type Installation,
} from '../src/config/commerce.ts';

const CASES: Array<{ equipment: Equipment; installation: Installation; expected: number }> = [
  { equipment: 'included', installation: 'none', expected: 2000 },
  { equipment: 'included', installation: 'onsite', expected: 2500 },
  { equipment: 'customer', installation: 'none', expected: 1000 },
  { equipment: 'customer', installation: 'onsite', expected: 1500 },
];

let failures = 0;

for (const testCase of CASES) {
  const { total } = quote(testCase);
  const passed = total === testCase.expected;
  if (!passed) failures += 1;

  const label = `${testCase.equipment.padEnd(8)} + instalación ${testCase.installation.padEnd(6)}`;
  console.log(`${passed ? 'PASS' : 'FAIL'}  ${label} → ${formatMXN(total)} (esperado ${formatMXN(testCase.expected)})`);
}

// Los mensajes deben arrastrar el importe calculado, no un literal duplicado.
const messageChecks: Array<[string, string, string]> = [
  ['equipo incluido', buildMessage({ equipment: 'included', installation: 'none' }), formatMXN(PRICING.equipmentIncluded)],
  ['equipo del cliente', buildMessage({ equipment: 'customer', installation: 'none' }), formatMXN(PRICING.customerEquipment)],
  ['instalación presencial', buildMessage({ equipment: 'included', installation: 'onsite' }), formatMXN(PRICING.equipmentIncluded + PRICING.onsiteInstallation)],
  ['vigencia de soporte', String(SUPPORT_MONTHS), '12'],
];

for (const [label, actual, expected] of messageChecks) {
  const passed = actual.includes(expected);
  if (!passed) failures += 1;
  console.log(`${passed ? 'PASS' : 'FAIL'}  mensaje ${label} contiene «${expected}»`);
}

// Los alias heredados deben derivarse de la fuente única, no declarar importes.
const legacy = await import('../src/config/pricing.ts');
const legacyChecks: Array<[string, number, number]> = [
  ['LEGACY_PRICING.ONN_CONFIGURED', legacy.LEGACY_PRICING.ONN_CONFIGURED, PRICING.equipmentIncluded],
  ['LEGACY_PRICING.CLIENT_EQUIPMENT', legacy.LEGACY_PRICING.CLIENT_EQUIPMENT, PRICING.customerEquipment],
  ['LEGACY_PRICING.INSTALLATION_PRESENCE', legacy.LEGACY_PRICING.INSTALLATION_PRESENCE, PRICING.onsiteInstallation],
  ['PRICING_TOTALS.onnInstall', legacy.PRICING_TOTALS.onnInstall, 2500],
  ['PRICING_TOTALS.clientInstall', legacy.PRICING_TOTALS.clientInstall, 1500],
];

for (const [label, actual, expected] of legacyChecks) {
  const passed = actual === expected;
  if (!passed) failures += 1;
  console.log(`${passed ? 'PASS' : 'FAIL'}  alias ${label} = ${actual} (esperado ${expected})`);
}

// Número comercial: si está configurado el enlace debe ser válido y llevar el mensaje;
// si no, debe ser null — nunca una URL a medias.
const link = whatsappLink();
const phoneOk = /^\d{12,13}$/.test(WHATSAPP_NUMBER);
if (!phoneOk) failures += 1;
console.log(`${phoneOk ? 'PASS' : 'FAIL'}  número comercial ${WHATSAPP_NUMBER} en formato wa.me (12–13 dígitos)`);

if (isWhatsAppConfigured()) {
  const urlOk = link !== null && /^https:\/\/wa\.me\/\d{10,15}\?text=/.test(link);
  const carriesTotal = link !== null && decodeURIComponent(link).includes(formatMXN(PRICING.equipmentIncluded));
  if (!urlOk || !carriesTotal) failures += 1;
  console.log(`${urlOk && carriesTotal ? 'PASS' : 'FAIL'}  whatsappLink() → ${urlOk ? 'URL válida' : 'URL malformada'}${carriesTotal ? ' + importe en el mensaje' : ' sin importe'}`);
  console.log(`INFO  número en pantalla: ${formatPhone()}`);
} else {
  const nullOk = link === null;
  if (!nullOk) failures += 1;
  console.log(`${nullOk ? 'PASS' : 'FAIL'}  sin número configurado whatsappLink() devuelve null`);
}

console.log(failures === 0 ? '\nOK: configuración comercial conforme al PRD §12.1' : `\n${failures} comprobación(es) fallaron`);
process.exit(failures === 0 ? 0 : 1);
