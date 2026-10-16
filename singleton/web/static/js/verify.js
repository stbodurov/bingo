// Hosted consent: build the proof on-device, submit it, return the result to the app.
(() => {
  const W = window.SingletonWallet;
  const el = document.getElementById('consent');
  const body = document.getElementById('v-body');
  const S = { id: el.dataset.session, scope: el.dataset.scope, challenge: el.dataset.challenge, required: JSON.parse(el.dataset.required), status: el.dataset.status, origin: el.dataset.origin, returnUrl: el.dataset.return, app: el.dataset.app };
  const popup = new URLSearchParams(location.search).get('mode') === 'popup' && window.opener;
  const h = (tag, cls, ...kids) => { const n = document.createElement(tag); if (cls) n.className = cls; for (const k of kids) n.append(k.nodeType ? k : document.createTextNode(k)); return n; };
  const btn = (cls, text, onclick) => { const b = h('button', 'btn ' + cls, text); b.type = 'button'; b.onclick = onclick; return b; };
  const label = { human: 'Unique human', human_doc: 'Document-verified human', age_over_13: 'Over 13', age_over_16: 'Over 16', age_over_18: 'Over 18', age_over_21: 'Over 21', region_us: 'Issued in US', region_gb: 'Issued in UK', region_au: 'Issued in AU', region_eu: 'Issued in EU' };

  function send(msg) {
    if (popup && S.origin) {
      try { window.opener.postMessage({ source: 'singleton', ...msg }, S.origin); } catch {}
    }
  }

  function finish(result) {
    send({ type: 'singleton:result', session: S.id, status: result.status, token: result.token, nullifier: result.nullifier, predicates: result.predicates, duplicateOf: result.duplicateOf || null, returning: result.returning });
    const dup = result.status === 'duplicate';
    body.replaceChildren(
      h('div', 'notice ' + (dup ? 'warn' : 'good'), dup ? `Verified — but this person already has an account at ${S.app}. ${S.app} decides what happens next.` : `Verified. ${S.app} now knows you're a real person — and nothing else.`),
      h('p', 'tiny muted mt16 mono', 'your pseudonym at ' + S.app + ': ' + result.nullifier.slice(0, 4) + '·' + result.nullifier.slice(4, 8) + '·' + result.nullifier.slice(8, 12)),
    );
    if (popup) setTimeout(() => window.close(), 1400);
    else if (S.returnUrl) setTimeout(() => { const u = new URL(S.returnUrl, location.origin); if (u.pathname.endsWith('/')) u.pathname += S.id; u.searchParams.set('singleton_session', S.id); location.href = u.toString(); }, 1400);
  }

  async function run() {
    if (S.status !== 'pending') {
      body.replaceChildren(h('div', 'notice warn', `This verification request is ${S.status}. Go back to ${S.app} and try again.`));
      return;
    }
    const pass = W.load();
    if (!pass) {
      const go = btn('btn-lime btn-lg', 'Create a Singleton Pass', () => (location.href = '/pass?return=' + encodeURIComponent(location.pathname + location.search)));
      go.style.width = '100%';
      body.replaceChildren(h('p', 'soft', 'You don’t have a Singleton Pass on this device yet. It takes about a minute, and you only do it once.'), h('div', 'mt16', go));
      return;
    }
    let enr = null;
    try { enr = await W.api('/api/enroll/' + pass.pubkey); pass.enrollment = enr; W.save(pass); } catch {}
    if (!enr || enr.status !== 'active') {
      const go = btn('btn-lime', 'Finish verifying my Pass', () => (location.href = '/pass?return=' + encodeURIComponent(location.pathname + location.search)));
      body.replaceChildren(h('div', 'notice warn', enr ? 'This Pass has been revoked.' : 'Your Pass isn’t verified yet.'), h('div', 'mt16', go));
      return;
    }
    const missing = S.required.filter((r) => !(enr.sets || []).includes(r));
    if (missing.length) {
      const msg = missing.map((m) => label[m] || m).join(', ');
      const upcoming = missing.map((m) => enr.upcoming?.[m]).filter(Boolean);
      body.replaceChildren(h('div', 'notice bad', `Your Pass can't prove: ${msg}.` + (upcoming.length ? ` You'll qualify on ${upcoming.join(', ')}.` : '')), h('div', 'row mt16', btn('', 'Decline', decline)));
      send({ type: 'singleton:unqualified', session: S.id, missing });
      return;
    }
    const steps = h('div', 'progress-steps');
    const approve = btn('btn-lime btn-lg', 'Prove it', async () => {
      approve.classList.add('busy');
      steps.replaceChildren();
      const step = (t) => { steps.querySelectorAll('div').forEach((d) => (d.className = 'done')); steps.append(h('div', 'on', t)); };
      try {
        const seedHex = await W.seed(async () => prompt('Passphrase for your Singleton Pass') || '');
        const t0 = performance.now();
        const { presentation, ms, ringSizes } = await W.present(seedHex, pass.pubkey, S.scope, S.challenge, S.required, (kind, set) => {
          if (kind === 'epoch') step('→ fetching the current anchored epoch');
          if (kind === 'bucket') step(`→ fetching your anonymity bucket for "${label[set] || set}"`);
          if (kind === 'sign') step('→ signing a linkable ring signature on this device');
        });
        step(`✓ proof built in ${ms} ms over ${ringSizes.join(' / ')} humans`);
        const res = await W.api(`/api/verify/${S.id}/present`, presentation);
        steps.querySelectorAll('div').forEach((d) => (d.className = 'done'));
        pass.history = [...(pass.history || []), { app: S.app, scope: S.scope, nullifier: res.nullifier, predicates: res.predicates, at: new Date().toISOString(), totalMs: Math.round(performance.now() - t0) }].slice(-100);
        W.save(pass);
        finish(res);
      } catch (ex) {
        steps.append(h('div', '', '✕ ' + ex.message));
        approve.classList.remove('busy');
      }
    });
    approve.style.flex = '1';
    body.replaceChildren(h('div', 'list', h('div', 'li', h('div', 'ico', '🪪'), h('div', 'grow', h('b', '', 'Pass ' + pass.fingerprint), h('small', '', (enr.sets || []).map((s) => label[s] || s).join(' · '))))), steps, h('div', 'row mt16', btn('', 'Decline', decline), approve));
  }

  function decline() {
    send({ type: 'singleton:cancel', session: S.id });
    body.replaceChildren(h('div', 'notice', 'You declined. Nothing was shared.'));
    if (popup) setTimeout(() => window.close(), 800);
  }

  run().catch((e) => body.replaceChildren(h('div', 'notice bad', e.message)));
})();
