// Console behaviours: mobile nav, copy buttons, confirmations, printing, chart tooltips, auth tabs.
(() => {
  document.getElementById('c-burger')?.addEventListener('click', () => document.getElementById('shell').classList.toggle('open'));
  document.querySelectorAll('[data-copy]').forEach((b) => b.addEventListener('click', () => {
    const t = document.querySelector(b.dataset.copy)?.textContent || '';
    navigator.clipboard?.writeText(t).then(() => { b.textContent = 'Copied'; setTimeout(() => (b.textContent = 'Copy'), 1500); });
  }));
  document.querySelectorAll('[data-confirm]').forEach((b) => b.addEventListener('click', (e) => { if (!confirm(b.dataset.confirm)) e.preventDefault(); }));
  document.querySelectorAll('[data-print]').forEach((b) => b.addEventListener('click', () => window.print()));
  document.querySelectorAll('[data-auth-tab]').forEach((b) => b.addEventListener('click', () => {
    document.querySelectorAll('[data-auth-tab]').forEach((x) => x.classList.toggle('on', x === b));
    document.querySelectorAll('[data-auth]').forEach((f) => f.classList.toggle('hidden', f.dataset.auth !== b.dataset.authTab));
  }));
  // Rich tooltip for stacked columns (the <title> keeps a native fallback).
  document.querySelectorAll('[data-chart]').forEach((wrap) => {
    const tip = document.createElement('div');
    tip.className = 'chart-tip'; tip.hidden = true; wrap.appendChild(tip);
    wrap.querySelectorAll('.col').forEach((g) => {
      g.addEventListener('pointerenter', () => {
        const [day, rest] = g.dataset.tip.split(' — ');
        tip.replaceChildren();
        const t = document.createElement('div'); t.className = 'muted tiny'; t.textContent = day; tip.append(t);
        rest.split(' · ').forEach((part) => { const d = document.createElement('div'); d.textContent = part; tip.append(d); });
        const box = g.getBBox ? g.querySelector('.hit').getBoundingClientRect() : null;
        const wr = wrap.getBoundingClientRect();
        if (box) { tip.style.left = box.left - wr.left + box.width / 2 + 'px'; tip.style.top = '40px'; }
        tip.hidden = false;
      });
      g.addEventListener('pointerleave', () => (tip.hidden = true));
    });
  });
})();
