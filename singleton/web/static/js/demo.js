// Demo relying parties: each uses the public SDK, then posts the signed token to its own
// backend (/demo/<app>/api/claim), which verifies it like a customer server would.
(() => {
  const root = document.getElementById('demo');
  if (!root) return;
  const pk = root.dataset.pk, appKey = root.dataset.app, action = root.dataset.act || '';
  const required = JSON.parse(root.dataset.required || '["human"]');
  const out = document.getElementById('demo-out');
  const say = (cls, text, sub) => {
    out.replaceChildren();
    const n = document.createElement('div'); n.className = 'notice ' + cls; n.textContent = text; out.append(n);
    if (sub) { const s = document.createElement('p'); s.className = 'tiny muted mt8 mono'; s.textContent = sub; out.append(s); }
  };
  document.querySelectorAll('[data-go]').forEach((b) => b.addEventListener('click', async () => {
    const valueEl = document.querySelector(b.dataset.value || '#value');
    const value = b.dataset.choice || (valueEl ? valueEl.value : '');
    if (appKey === 'driftwood' && !/^[A-Za-z0-9_]{3,20}$/.test(value)) return say('bad', 'Pick a handle: 3–20 letters, numbers or underscores.');
    b.classList.add('busy');
    try {
      const r = await window.Singleton.verify({ publishableKey: pk, required, action, reference: appKey === 'driftwood' ? value : '' });
      const res = await fetch(`/demo/${appKey}/api/claim`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ token: r.token, value }) });
      const data = await res.json();
      const p = 'pseudonym ' + (data.nullifier || r.nullifier || '').slice(0, 12);
      if (!res.ok) return say('warn', data.error, p); // keep the explanation on screen
      if (appKey === 'driftwood') say('good', `Welcome to Driftwood, @${data.value}! Verified human, 16+.`, p + ' · now try signing up again with another handle');
      else if (appKey === 'frontrow') say('good', `Ticket claimed — seat ${data.value}. Enjoy the show.`, p + ' · try claiming a second one');
      else say('good', 'Vote counted. Nobody — not even us — knows how you voted.', p);
      setTimeout(() => location.reload(), 5000); // show the new member in the list
    } catch (e) {
      say(e.code === 'unqualified' ? 'warn' : 'bad', e.message);
    } finally { b.classList.remove('busy'); }
  }));
})();
