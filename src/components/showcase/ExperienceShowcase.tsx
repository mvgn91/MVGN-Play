import { useRef, useState, type KeyboardEvent } from 'react';

type TabKey = 'interfaz' | 'aplicaciones' | 'soporte';

const TABS: ReadonlyArray<{ key: TabKey; label: string; hint: string }> = [
  { key: 'interfaz', label: 'Interfaz', hint: 'Pantalla de inicio organizada' },
  { key: 'aplicaciones', label: 'Aplicaciones', hint: 'Películas, series y TV' },
  { key: 'soporte', label: 'Soporte', hint: 'Diagnóstico y sustitución' },
];

const APPS = [
  { group: 'Películas y series', items: ['Aplicaciones incluidas en la configuración'] },
  { group: 'Televisión', items: ['Canales y guía dentro de las aplicaciones'] },
  { group: 'Utilidades', items: ['Lanzador personalizado', 'Ajustes del sistema'] },
];

const LAUNCHER_TILES = ['Inicio', 'Apps', 'TV', 'Ajustes', 'Guía', 'Perfil'];

const SUPPORT_STEPS = [
  { step: '01', title: 'Diagnóstico', body: 'Revisamos la configuración, la red y el comportamiento de las aplicaciones utilizadas.' },
  { step: '02', title: 'Ajuste o sustitución', body: 'Corregimos la configuración o, si la aplicación dejó de servir, instalamos una alternativa viable.' },
  { step: '03', title: 'Seguimiento', body: 'Queda constancia del cambio y de lo que cubre el periodo de soporte.' },
];

export default function ExperienceShowcase() {
  const [active, setActive] = useState<TabKey>('interfaz');
  const tabRefs = useRef<Array<HTMLButtonElement | null>>([]);

  const focusTab = (index: number) => {
    const target = tabRefs.current[index];
    if (target) {
      target.focus();
      setActive(TABS[index].key);
    }
  };

  const onKeyDown = (event: KeyboardEvent<HTMLButtonElement>, index: number) => {
    if (event.key === 'ArrowRight') {
      event.preventDefault();
      focusTab((index + 1) % TABS.length);
    } else if (event.key === 'ArrowLeft') {
      event.preventDefault();
      focusTab((index - 1 + TABS.length) % TABS.length);
    } else if (event.key === 'Home') {
      event.preventDefault();
      focusTab(0);
    } else if (event.key === 'End') {
      event.preventDefault();
      focusTab(TABS.length - 1);
    }
  };

  return (
    <div className="grid gap-8 lg:grid-cols-[16rem_1fr] lg:items-start">
      <div role="tablist" aria-label="Estados de la experiencia" className="grid grid-cols-3 gap-2 lg:grid-cols-1">
        {TABS.map((tab, index) => {
          const selected = tab.key === active;
          return (
            <button
              key={tab.key}
              ref={(node) => {
                tabRefs.current[index] = node;
              }}
              type="button"
              role="tab"
              id={`tab-${tab.key}`}
              aria-selected={selected}
              aria-controls={`panel-${tab.key}`}
              tabIndex={selected ? 0 : -1}
              onClick={() => setActive(tab.key)}
              onKeyDown={(event) => onKeyDown(event, index)}
              className={[
                'flex min-w-0 flex-col items-start gap-1 rounded-none border px-3 py-2.5 text-left transition-colors sm:px-4 sm:py-3',
                selected
                  ? 'border-crimson/70 bg-surface text-ink'
                  : 'border-line bg-surface/40 text-ink-muted hover:border-ink/30 hover:text-ink',
              ].join(' ')}
            >
              <span className="font-display text-xs font-semibold sm:text-sm">{tab.label}</span>
              <span className="hidden text-xs text-ink-muted sm:block">{tab.hint}</span>
            </button>
          );
        })}
      </div>

      <div className="rounded-panel border border-line bg-surface/60 p-5 sm:p-8">
        <div
          role="tabpanel"
          id="panel-interfaz"
          aria-labelledby="tab-interfaz"
          hidden={active !== 'interfaz'}
          tabIndex={0}
          className="focus-visible:outline-none"
        >
          <div className="rounded-none border border-line bg-bg p-4">
            <div className="mb-3 flex items-center justify-between">
              <span className="label-tech">Inicio</span>
              <span className="font-mono text-[0.625rem] text-cool">configuración personalizada</span>
            </div>
            <div className="grid grid-cols-3 gap-2 sm:grid-cols-6">
              {LAUNCHER_TILES.map((tile, index) => (
                <div
                  key={tile}
                  className={[
                    'flex flex-col gap-1.5 rounded-none border p-2',
                    index === 0 ? 'border-crimson/70 bg-surface/80' : 'border-line/70 bg-surface/60',
                  ].join(' ')}
                >
                  <div aria-hidden="true" className={['h-8 rounded-none', index === 0 ? 'bg-crimson/25' : 'bg-surface-elevated'].join(' ')} />
                  <span className="truncate text-[0.6875rem] text-ink-muted">{tile}</span>
                </div>
              ))}
            </div>
          </div>
          <p className="mt-4 text-sm text-ink-muted">
            La pantalla de inicio se organiza según lo que realmente usas: accesos directos en el orden acordado, sin
            relleno ni aplicaciones que no hayas pedido.
          </p>
        </div>

        <div
          role="tabpanel"
          id="panel-aplicaciones"
          aria-labelledby="tab-aplicaciones"
          hidden={active !== 'aplicaciones'}
          tabIndex={0}
          className="focus-visible:outline-none"
        >
          <dl className="grid gap-4 sm:grid-cols-3">
            {APPS.map((group) => (
              <div key={group.group} className="rounded-none border border-line bg-bg p-4">
                <dt className="label-tech">{group.group}</dt>
                <dd className="mt-2 space-y-1 text-sm text-ink-muted">
                  {group.items.map((item) => (
                    <p key={item}>{item}</p>
                  ))}
                </dd>
              </div>
            ))}
          </dl>
          <p className="mt-4 text-sm text-ink-muted">
            Las aplicaciones incluidas son el corazón del servicio: se entregan configuradas y probadas para ver
            películas, series y televisión. Si una deja de funcionar o cambia, el soporte la ajusta o instala una
            alternativa viable.
          </p>
        </div>

        <div
          role="tabpanel"
          id="panel-soporte"
          aria-labelledby="tab-soporte"
          hidden={active !== 'soporte'}
          tabIndex={0}
          className="focus-visible:outline-none"
        >
          <ol className="grid gap-3 sm:grid-cols-3">
            {SUPPORT_STEPS.map((item) => (
              <li key={item.step} className="rounded-none border border-line bg-bg p-4">
                <span className="font-mono text-xs text-cool">{item.step}</span>
                <h3 className="mt-2 font-display text-base font-semibold">{item.title}</h3>
                <p className="mt-1 text-sm text-ink-muted">{item.body}</p>
              </li>
            ))}
          </ol>
          <p className="mt-4 text-sm text-ink-muted">
            Durante los 12 meses el soporte se ocupa de mantener funcionando las aplicaciones incluidas, con las
            sustituciones que hagan falta. No incluye reparaciones físicas, reemplazo de hardware ni licencias de
            terceros.
          </p>
        </div>
      </div>
    </div>
  );
}
