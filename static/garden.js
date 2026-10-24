(() => {
  document.addEventListener('DOMContentLoaded', () => {
    const themeButton = document.querySelector('.theme-control');
    const canvas = document.querySelector('.rain-field');
    const context = canvas.getContext('2d');
    if (!context) return;
    const scene = canvas.parentElement;
    const motionButton = document.querySelector('.motion-control');
    const reducedMotion = matchMedia('(prefers-reduced-motion: reduce)');
    let paused = reducedMotion.matches;
    let frame = 0;
    let width = 0;
    let height = 0;
    let pointer = 0;
    let drift = 0;
    let focus = '';
    let pull = 0;
    let phase = 0;
    let lastTime = 0;
    const drop = new Image();
    drop.addEventListener('load', requestDraw);
    drop.src = '/static/images/garden-drop.webp';
    const particles = Array.from({ length: 6 }, (_, index) => ({
      offset: index / 6,
      lane: ((index * 7) % 11 - 5) / 5,
    }));

    function draw(time) {
      frame = 0;
      const active = !paused && !document.hidden;
      const elapsed = lastTime ? Math.min(time - lastTime, 50) : 0;
      lastTime = time;
      if (active) {
        phase += elapsed * .00007;
        drift += (pointer - drift) * .035;
        const targetPull = focus === 'left' ? -1 : focus === 'right' ? 1 : 0;
        pull += (targetPull - pull) * .035;
      }
      if (reducedMotion.matches) {
        drift = 0;
        pull = 0;
      }
      context.clearRect(0, 0, width, height);
      for (const particle of particles) {
        if (!drop.complete || !drop.naturalWidth) continue;
        const progress = (particle.offset + phase) % 1;
        const curve = Math.sin(progress * Math.PI) * (particle.lane * .012 + drift * .012 + pull * .035);
        const x = width * (.710 + particle.lane * .045 + curve);
        const y = height * (.315 + progress * .18);
        const size = width * .022;
        context.globalAlpha = Math.sin(progress * Math.PI) * .7 + .08;
        context.save();
        context.translate(x, y);
        context.rotate(curve * 1.8);
        context.drawImage(drop, -size / 2, -size / 2, size, size);
        context.restore();
      }
      context.globalAlpha = 1;
      if (active) frame = requestAnimationFrame(draw);
    }

    function requestDraw() {
      if (!frame) frame = requestAnimationFrame(draw);
    }
    function resize() {
      const bounds = scene.getBoundingClientRect();
      width = bounds.width;
      height = bounds.height;
      const density = Math.min(devicePixelRatio || 1, 2);
      canvas.width = Math.round(width * density);
      canvas.height = Math.round(height * density);
      context.setTransform(density, 0, 0, density, 0, 0);
      requestDraw();
    }
    function syncMotion() {
      motionButton.hidden = reducedMotion.matches;
      motionButton.setAttribute('aria-pressed', String(paused));
      const label = paused ? motionButton.dataset.resumeLabel : motionButton.dataset.pauseLabel;
      motionButton.setAttribute('aria-label', label);
      motionButton.querySelector('span').textContent = label;
      lastTime = 0;
      requestDraw();
    }
    motionButton.addEventListener('click', () => { paused = !paused; syncMotion(); });
    reducedMotion.addEventListener('change', () => { paused = reducedMotion.matches; syncMotion(); });
    themeButton.addEventListener('click', requestDraw);
    document.addEventListener('visibilitychange', () => { lastTime = 0; requestDraw(); });
    document.querySelector('.garden-main').addEventListener('pointermove', event => {
      if (reducedMotion.matches || event.pointerType === 'touch') return;
      pointer = Math.max(-1, Math.min(1, (event.clientX / innerWidth - .5) * 2));
    }, { passive: true });
    document.querySelector('.garden-main').addEventListener('pointerleave', () => { pointer = 0; });
    for (const link of document.querySelectorAll('[data-rain]')) {
      const emphasize = () => { focus = link.dataset.rain; requestDraw(); };
      const reset = () => { focus = ''; requestDraw(); };
      link.addEventListener('pointerenter', emphasize);
      link.addEventListener('focus', emphasize);
      link.addEventListener('pointerleave', reset);
      link.addEventListener('blur', reset);
    }
    if ('ResizeObserver' in window) new ResizeObserver(resize).observe(scene);
    else window.addEventListener('resize', resize, { passive: true });
    resize();
    syncMotion();
  });
})();
