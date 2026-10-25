import { useEffect, useState } from 'react';
import CrystalizedBall from '@/components/CrystalizedBall';

/**
 * Adaptador del CrystalizedBall (React Bits) al sistema de esta web:
 * color tomado del token `--crimson`, tema sincronizado con `data-theme`,
 * pausado con `prefers-reduced-motion` y menos partículas en pantallas chicas.
 */
export default function BallHero() {
  const [theme, setTheme] = useState<'dark' | 'light'>('dark');
  const [color, setColor] = useState('#dc143c');
  const [paused, setPaused] = useState(false);
  const [particleCount, setParticleCount] = useState(14000);

  useEffect(() => {
    const root = document.documentElement;

    const syncTheme = () => {
      const next = root.dataset.theme === 'light' ? 'light' : 'dark';
      setTheme(next);
      const token = getComputedStyle(root).getPropertyValue('--crimson').trim();
      if (token) setColor(token);
    };

    syncTheme();

    const observer = new MutationObserver(syncTheme);
    observer.observe(root, { attributes: true, attributeFilter: ['data-theme'] });

    const reduced = window.matchMedia('(prefers-reduced-motion: reduce)');
    const syncReduced = () => setPaused(reduced.matches);
    syncReduced();
    reduced.addEventListener('change', syncReduced);

    const small = window.matchMedia('(max-width: 639px)');
    const syncSmall = () => setParticleCount(small.matches ? 6000 : 14000);
    syncSmall();
    small.addEventListener('change', syncSmall);

    return () => {
      observer.disconnect();
      reduced.removeEventListener('change', syncReduced);
      small.removeEventListener('change', syncSmall);
    };
  }, []);

  return (
    <div aria-hidden="true" className="relative mx-auto aspect-square w-full max-w-[20rem] sm:max-w-[24rem] lg:max-w-[26rem]">
      <div
        className="pointer-events-none absolute inset-[10%] rounded-none opacity-70"
        style={{ background: 'radial-gradient(circle, color-mix(in oklab, var(--crimson) 30%, transparent), transparent 68%)' }}
      />
      <CrystalizedBall
        color={color}
        theme={theme}
        size={0.92}
        speed={1}
        paused={paused}
        intro={false}
        interactive
        particleCount={particleCount}
        className="relative h-full w-full"
      />
    </div>
  );
}
