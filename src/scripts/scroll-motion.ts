/**
 * Movimiento al hacer scroll: revelado de secciones + parallax ligero.
 * Vanilla, un solo loop por frame, sin librerías. Si el usuario pide
 * movimiento reducido, todo queda visible y estático.
 */
const root = document.documentElement;
const reduced = window.matchMedia('(prefers-reduced-motion: reduce)');

let teardown: (() => void) | null = null;

function showEverything() {
  document.querySelectorAll<HTMLElement>('[data-reveal]').forEach((element) => {
    element.classList.add('is-in');
  });
}

function enable() {
  teardown?.();
  root.dataset.motion = 'on';

  const reveals = Array.from(document.querySelectorAll<HTMLElement>('[data-reveal]'));
  const parallax = Array.from(document.querySelectorAll<HTMLElement>('[data-parallax]'));

  // Revelado dinámico: entra en vista → aparece; sale de vista → se va, y
  // al volver a entrar vuelve a animarse. Los elementos siguen observados.
  if ('IntersectionObserver' in window && reveals.length > 0) {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          const element = entry.target as HTMLElement;
          element.classList.toggle('is-in', entry.isIntersecting);
        });
      },
      { rootMargin: '0px 0px -6% 0px', threshold: 0.1 },
    );

    reveals.forEach((element) => observer.observe(element));
    teardown = () => observer.disconnect();
  } else {
    showEverything();
  }

  if (parallax.length === 0) return;

  let queued = false;

  const paint = () => {
    queued = false;
    const viewport = window.innerHeight;

    parallax.forEach((element) => {
      const speed = Number.parseFloat(element.dataset.parallax ?? '0.12');
      const rect = element.getBoundingClientRect();
      if (rect.bottom < -200 || rect.top > viewport + 200) return;

      const center = rect.top + rect.height / 2;
      const progress = (center - viewport / 2) / viewport;
      const shift = Math.max(-90, Math.min(90, -progress * speed * 170));
      element.style.setProperty('--py', `${shift.toFixed(1)}px`);
    });
  };

  const schedule = () => {
    if (queued) return;
    queued = true;
    requestAnimationFrame(paint);
  };

  paint();
  window.addEventListener('scroll', schedule, { passive: true });
  window.addEventListener('resize', schedule, { passive: true });

  const previous = teardown;
  teardown = () => {
    previous?.();
    window.removeEventListener('scroll', schedule);
    window.removeEventListener('resize', schedule);
    parallax.forEach((element) => element.style.removeProperty('--py'));
  };
}

function disable() {
  teardown?.();
  teardown = null;
  root.dataset.motion = 'off';
  showEverything();
}

if (reduced.matches) {
  disable();
} else {
  enable();
}

reduced.addEventListener('change', (event) => {
  if (event.matches) disable();
  else enable();
});
