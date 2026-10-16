// Landing page: the anonymity-set field, code tabs, and the proof-of-work waitlist.
(() => {
  // ---------------------------------------------------------------- dot field
  const canvas = document.getElementById('field');
  if (canvas) {
    const ctx = canvas.getContext('2d');
    const reduce = matchMedia('(prefers-reduced-motion: reduce)').matches;
    let dots = [];
    let w = 0, h = 0, dpr = 1, me = null, mouse = { x: -999, y: -999 }, t0 = performance.now();
    const build = () => {
      const r = canvas.getBoundingClientRect();
      dpr = Math.min(2, window.devicePixelRatio || 1);
      w = r.width; h = r.height;
      canvas.width = w * dpr; canvas.height = h * dpr;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      const gap = w < 700 ? 22 : 26;
      dots = [];
      for (let y = gap / 2; y < h; y += gap) {
        for (let x = gap / 2; x < w; x += gap) {
          dots.push({ x: x + (Math.random() - 0.5) * 8, y: y + (Math.random() - 0.5) * 8, p: Math.random() * Math.PI * 2, s: 0.4 + Math.random() * 0.6 });
        }
      }
      // "you": a dot to the right, behind the proof card's edge on desktop
      // Desktop: in the gutter between headline and card. Mobile: top-right corner.
      const cand = w > 960
        ? dots.filter((d) => d.x > w * 0.5 && d.x < w * 0.56 && d.y > h * 0.3 && d.y < h * 0.7)
        : dots.filter((d) => d.x > w * 0.78 && d.y > h * 0.06 && d.y < h * 0.12);
      me = cand[Math.floor(cand.length / 2)] || dots[Math.floor(dots.length / 2)];
    };
    const draw = (now) => {
      const t = (now - t0) / 1000;
      ctx.clearRect(0, 0, w, h);
      for (const d of dots) {
        const dx = d.x - mouse.x, dy = d.y - mouse.y;
        const near = Math.max(0, 1 - Math.sqrt(dx * dx + dy * dy) / 140);
        const a = 0.12 + 0.1 * Math.sin(t * d.s + d.p) + near * 0.45;
        ctx.fillStyle = `rgba(243,243,241,${a.toFixed(3)})`;
        ctx.beginPath();
        ctx.arc(d.x, d.y, 1.1 + near * 0.9, 0, Math.PI * 2);
        ctx.fill();
      }
      if (me) {
        const pulse = (t % 2.4) / 2.4;
        ctx.strokeStyle = `rgba(212,255,79,${(0.55 * (1 - pulse)).toFixed(3)})`;
        ctx.lineWidth = 1.2;
        ctx.beginPath();
        ctx.arc(me.x, me.y, 6 + pulse * 34, 0, Math.PI * 2);
        ctx.stroke();
        ctx.fillStyle = '#d4ff4f';
        ctx.shadowColor = '#d4ff4f';
        ctx.shadowBlur = 16;
        ctx.beginPath();
        ctx.arc(me.x, me.y, 3.6, 0, Math.PI * 2);
        ctx.fill();
        ctx.shadowBlur = 0;
      }
      if (!reduce) requestAnimationFrame(draw);
    };
    build();
    requestAnimationFrame(draw);
    let rt;
    window.addEventListener('resize', () => { clearTimeout(rt); rt = setTimeout(build, 150); });
    canvas.parentElement.addEventListener('pointermove', (e) => {
      const r = canvas.getBoundingClientRect();
      mouse = { x: e.clientX - r.left, y: e.clientY - r.top };
    });
    canvas.parentElement.addEventListener('pointerleave', () => (mouse = { x: -999, y: -999 }));
  }

  // ---------------------------------------------------------------- code tabs
  const code = document.getElementById('code');
  code?.querySelectorAll('[data-tab]').forEach((b) => b.addEventListener('click', () => {
    code.querySelectorAll('[data-tab]').forEach((x) => x.classList.toggle('on', x === b));
    code.querySelectorAll('[data-pane]').forEach((p) => (p.hidden = p.dataset.pane !== b.dataset.tab));
  }));

  // ---------------------------------------------------------------- waitlist
  const form = document.getElementById('wl');
  if (!form) return;
  const err = document.getElementById('wl-err');
  const powEl = document.getElementById('pow');
  const btn = document.getElementById('wl-btn');
  const resultEl = document.getElementById('wl-result');
  let pow = null;           // { challenge, sig, nonce }
  let powPromise = null;

  const setTrack = (track) => {
    form.track.value = track;
    document.querySelectorAll('[data-track-btn]').forEach((b) => b.classList.toggle('on', b.dataset.trackBtn === track));
    form.querySelector('.platform-only').classList.toggle('hidden', track !== 'platform');
    btn.firstChild.textContent = track === 'platform' ? 'Request design-partner access ' : 'Join the waitlist ';
  };
  document.querySelectorAll('[data-track-btn]').forEach((b) => b.addEventListener('click', () => setTrack(b.dataset.trackBtn)));
  document.querySelectorAll('[data-track]').forEach((a) => a.addEventListener('click', () => setTrack(a.dataset.track)));
  if (location.hash === '#join-platform') setTrack('platform');

  const setPow = (state, text) => {
    powEl.className = 'pow ' + state;
    powEl.querySelector('span').textContent = text;
  };

  const solve = () => {
    if (powPromise) return powPromise;
    powPromise = (async () => {
      setPow('working', 'Proving you’re not a bot (proof-of-work)…');
      const ch = await (await fetch('/api/waitlist/challenge')).json();
      const nonce = await new Promise((resolve, reject) => {
        const worker = new Worker('/static/js/pow-worker.js');
        worker.onmessage = (e) => {
          if (e.data.done) { worker.terminate(); setPow('done', `Proof-of-work done · ${e.data.tries.toLocaleString()} hashes in ${e.data.ms} ms`); resolve(e.data.nonce); }
        };
        worker.onerror = (e) => { worker.terminate(); reject(e); };
        worker.postMessage({ challenge: ch.challenge, bits: ch.bits });
      });
      pow = { challenge: ch.challenge, sig: ch.sig, nonce };
      return pow;
    })().catch((e) => { powPromise = null; setPow('', 'Could not compute proof-of-work — retry'); throw e; });
    return powPromise;
  };
  // Start the work as soon as someone shows intent, so submit feels instant.
  form.addEventListener('focusin', () => solve().catch(() => {}), { once: true });

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    err.textContent = '';
    const fd = Object.fromEntries(new FormData(form));
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/.test(fd.email || '')) { err.textContent = 'Enter a valid email.'; return; }
    btn.classList.add('busy');
    try {
      const p = await solve();
      const res = await fetch('/api/waitlist', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ ...fd, pow: p }) });
      const data = await res.json();
      powPromise = null; pow = null; // challenges are single-use
      if (!res.ok) throw new Error(data.error || 'Something went wrong');
      try { localStorage.setItem('singleton:waitlist', data.code); } catch {}
      showResult(data);
    } catch (ex) {
      err.textContent = ex.message;
    } finally {
      btn.classList.remove('busy');
    }
  });

  function showResult(d) {
    form.classList.add('hidden');
    document.querySelector('.tracks').classList.add('hidden');
    resultEl.classList.remove('hidden');
    const person = d.track === 'person';
    resultEl.innerHTML = '';
    const el = (tag, cls, text) => { const n = document.createElement(tag); if (cls) n.className = cls; if (text !== undefined) n.textContent = text; return n; };
    resultEl.append(
      el('span', 'eyebrow', d.existing ? 'Welcome back' : 'You’re in line'),
      el('div', 'pos', '#' + Number(d.position).toLocaleString()),
      el('p', 'soft', `${person ? 'of' : 'platforms in queue:'} ${Number(d.total).toLocaleString()} · ${d.tier}`),
    );
    if (person) {
      resultEl.append(el('p', 'mt16', 'Move up 5 places for every friend who joins with your link. 3 friends = early access. 10 = Founding Human.'));
      const share = el('div', 'share');
      const input = el('input');
      input.value = d.shareUrl; input.readOnly = true; input.setAttribute('aria-label', 'Your referral link');
      const copy = el('button', 'btn btn-sm', 'Copy');
      copy.type = 'button';
      copy.onclick = () => navigator.clipboard?.writeText(d.shareUrl).then(() => (copy.textContent = 'Copied'));
      share.append(input, copy);
      resultEl.append(share);
      const x = el('a', 'btn btn-sm mt16', 'Share on X');
      x.href = 'https://twitter.com/intent/tweet?text=' + encodeURIComponent('I just joined the waitlist for Singleton — prove you’re human without giving up your identity.') + '&url=' + encodeURIComponent(d.shareUrl);
      x.target = '_blank'; x.rel = 'noopener';
      resultEl.append(x);
    } else {
      resultEl.append(el('p', 'mt16', 'We review platform requests daily and invite design partners in waves. In the meantime, the console sandbox is open.'));
      const c = el('a', 'btn btn-lime btn-sm mt16', 'Open the console sandbox →');
      c.href = '/console/login?mode=signup';
      resultEl.append(c);
    }
  }

  // Returning visitors see their spot.
  try {
    const code = localStorage.getItem('singleton:waitlist');
    if (code) fetch('/api/waitlist/' + encodeURIComponent(code)).then((r) => (r.ok ? r.json() : null)).then((d) => {
      if (d) { const note = document.createElement('p'); note.className = 'small muted mt16'; note.textContent = `You're #${d.position.toLocaleString()} in the ${d.track === 'person' ? 'Pass' : 'platform'} queue.`; form.after(note); }
    });
  } catch {}
})();
