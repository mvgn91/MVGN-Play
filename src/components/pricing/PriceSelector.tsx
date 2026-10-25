import { useId, useState } from 'react';
import {
  CONTACT_FALLBACK_HREF,
  PRICING,
  SUPPORT_MONTHS,
  formatMXN,
  quote,
  whatsappLink,
  type Equipment,
  type Installation,
} from '@/config/commerce';

const INCLUDED_BY_EQUIPMENT: Record<Equipment, string[]> = {
  included: [
    'Dispositivo ONN 4K seleccionado para la oferta',
    'Configuración inicial y personalización de interfaz',
    'Aplicaciones incluidas para películas, series y televisión, con soporte que las mantiene funcionando',
    'Pruebas funcionales antes de la entrega',
    `Soporte técnico durante ${SUPPORT_MONTHS} meses`,
  ],
  customer: [
    'Evaluación básica de compatibilidad',
    'Configuración de la TV Box que ya tienes',
    'Personalización de interfaz y ajustes',
    'Aplicaciones incluidas para películas, series y televisión, con soporte que las mantiene funcionando',
    `Soporte técnico durante ${SUPPORT_MONTHS} meses`,
  ],
};

const SITE_REGION = 'Zona Metropolitana de Guadalajara';

const OPTION_CLASS =
  'flex cursor-pointer flex-col gap-1 rounded-none border px-4 py-3 transition-colors has-[:focus-visible]:outline has-[:focus-visible]:outline-2 has-[:focus-visible]:outline-offset-2 has-[:focus-visible]:outline-cool';

export default function PriceSelector() {
  const [equipment, setEquipment] = useState<Equipment>('included');
  const [installation, setInstallation] = useState<Installation>('none');

  const equipmentName = useId();
  const installName = useId();
  const totalId = useId();

  const current = quote({ equipment, installation });
  const link = whatsappLink({ equipment, installation });

  return (
    <div className="grid gap-6 lg:grid-cols-[1.15fr_0.85fr] lg:items-start">
      <div className="space-y-8">
        <fieldset>
          <legend className="label-tech">1 · Equipo</legend>
          <div className="mt-3 grid gap-3 sm:grid-cols-2">
            {(
              [
                {
                  value: 'included' as Equipment,
                  title: 'ONN 4K incluido',
                  price: PRICING.equipmentIncluded,
                  hint: 'Te entregamos el equipo configurado y probado.',
                },
                {
                  value: 'customer' as Equipment,
                  title: 'Ya tengo mi TV Box',
                  price: PRICING.customerEquipment,
                  hint: 'Configuramos tu equipo si es compatible y funcional.',
                },
              ] satisfies Array<{ value: Equipment; title: string; price: number; hint: string }>
            ).map((option) => {
              const selected = option.value === equipment;
              return (
                <label
                  key={option.value}
                  className={[
                    OPTION_CLASS,
                    selected ? 'border-crimson/70 bg-surface' : 'border-line bg-surface/40 hover:border-ink/30',
                  ].join(' ')}
                >
                  <span className="flex items-center gap-2">
                    <input
                      type="radio"
                      name={equipmentName}
                      value={option.value}
                      checked={selected}
                      onChange={() => setEquipment(option.value)}
                      className="accent-crimson"
                    />
                    <span className="font-display text-sm font-semibold">{option.title}</span>
                  </span>
                  <span className="font-mono text-xs text-cool">{formatMXN(option.price)}</span>
                  <span className="text-xs text-ink-muted">{option.hint}</span>
                </label>
              );
            })}
          </div>
        </fieldset>

        <fieldset>
          <legend className="label-tech">2 · Instalación</legend>
          <div className="mt-3 grid gap-3 sm:grid-cols-2">
            {(
              [
                {
                  value: 'none' as Installation,
                  title: 'Sin instalación presencial',
                  price: 0,
                  hint: 'Entrega coordinada en la ZMG.',
                },
                {
                  value: 'onsite' as Installation,
                  title: 'Instalación presencial',
                  price: PRICING.onsiteInstallation,
                  hint: 'Visita dentro de la zona de cobertura.',
                },
              ] satisfies Array<{ value: Installation; title: string; price: number; hint: string }>
            ).map((option) => {
              const selected = option.value === installation;
              return (
                <label
                  key={option.value}
                  className={[
                    OPTION_CLASS,
                    selected ? 'border-crimson/70 bg-surface' : 'border-line bg-surface/40 hover:border-ink/30',
                  ].join(' ')}
                >
                  <span className="flex items-center gap-2">
                    <input
                      type="radio"
                      name={installName}
                      value={option.value}
                      checked={selected}
                      onChange={() => setInstallation(option.value)}
                      className="accent-crimson"
                    />
                    <span className="font-display text-sm font-semibold">{option.title}</span>
                  </span>
                  <span className="font-mono text-xs text-cool">
                    {option.price === 0 ? 'incluido' : `+ ${formatMXN(option.price)}`}
                  </span>
                  <span className="text-xs text-ink-muted">{option.hint}</span>
                </label>
              );
            })}
          </div>
        </fieldset>

        <div>
          <p className="label-tech">Incluye</p>
          <ul className="mt-3 grid gap-2 text-sm text-ink-muted sm:grid-cols-2">
            {INCLUDED_BY_EQUIPMENT[equipment].map((item) => (
              <li key={item} className="flex gap-2">
                <span aria-hidden="true" className="mt-2 h-1 w-1 shrink-0 rounded-none bg-crimson" />
                {item}
              </li>
            ))}
          </ul>
        </div>
      </div>

      <div className="rounded-panel border border-line bg-surface/70 p-5 sm:p-6 lg:sticky lg:top-24">
        <p className="label-tech">Total estimado</p>
        <p id={totalId} aria-live="polite" className="mt-2 font-display text-4xl font-semibold">
          {formatMXN(current.total)}
        </p>

        <dl className="mt-5 space-y-2 border-t border-line pt-4 text-sm">
          {current.lines.map((line) => (
            <div key={line.label} className="flex items-baseline justify-between gap-4">
              <dt className="text-ink-muted">{line.label}</dt>
              <dd className="font-mono text-xs text-ink">{formatMXN(line.amount)}</dd>
            </div>
          ))}
          <div className="flex items-baseline justify-between gap-4 border-t border-line pt-2">
            <dt className="text-ink">Soporte</dt>
            <dd className="font-mono text-xs text-ink">{SUPPORT_MONTHS} meses incluidos</dd>
          </div>
        </dl>

        {link ? (
          <a
            href={link}
            target="_blank"
            rel="noopener noreferrer"
            className="mt-5 inline-flex w-full items-center justify-center gap-2 rounded-none bg-crimson px-5 py-3 text-sm font-medium text-white transition-colors hover:bg-crimson-deep"
          >
            Consultar esta opción por WhatsApp
          </a>
        ) : (
          <p className="mt-5 rounded-none border border-dashed border-line px-3 py-2 text-xs text-ink-muted">
            El enlace de WhatsApp se habilita cuando se defina el número comercial.{' '}
            <a className="text-cool underline-offset-2 hover:underline" href={CONTACT_FALLBACK_HREF}>
              Ver otros medios de contacto
            </a>
          </p>
        )}

        <p className="mt-4 text-xs text-ink-muted">
          Incluye las aplicaciones configuradas para películas, series y televisión y el soporte que las mantiene
          funcionando; su alcance puede cambiar con el tiempo. Envíos fuera de la {SITE_REGION} y traslados fuera de
          la zona de cobertura se cotizan por separado.
        </p>
      </div>
    </div>
  );
}
