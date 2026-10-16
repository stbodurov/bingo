// Singleton Pass — the holder's wallet. Views: welcome, enroll (5 paths), home (apps,
// agents, recovery, settings) and vouch-for-someone.
(() => {
  const W = window.SingletonWallet;
  const root = document.getElementById('pass-root');
  const ISSUERS = JSON.parse(root.dataset.issuers || '[]');
  const SETS = JSON.parse(root.dataset.sets || '[]');
  const setLabel = (n) => (SETS.find((s) => s.name === n) || { label: n }).label;
  const params = new URLSearchParams(location.search);
  const returnTo = params.get('return');

  // ------------------------------------------------------------ tiny DOM helper
  function h(tag, attrs = {}, ...kids) {
    const el = document.createElement(tag);
    for (const [k, v] of Object.entries(attrs || {})) {
      if (v === null || v === undefined || v === false) continue;
      if (k === 'class') el.className = v;
      else if (k.startsWith('on')) el.addEventListener(k.slice(2), v);
      else if (k === 'html') el.innerHTML = v;
      else el.setAttribute(k, v === true ? '' : v);
    }
    for (const k of kids.flat()) if (k !== null && k !== undefined && k !== false) el.append(k.nodeType ? k : document.createTextNode(String(k)));
    return el;
  }
  const show = (...nodes) => root.replaceChildren(...nodes);
  const ago = (iso) => { const s = (Date.now() - new Date(iso)) / 1000; return s < 60 ? 'just now' : s < 3600 ? `${Math.round(s / 60)}m ago` : s < 86400 ? `${Math.round(s / 3600)}h ago` : `${Math.round(s / 86400)}d ago`; };
  const pseud = (n) => (n ? `${n.slice(0, 4)}·${n.slice(4, 8)}·${n.slice(8, 12)}` : '—');
  const busy = (btn, on) => btn && btn.classList.toggle('busy', on);
  const askPass = () => new Promise((resolve) => sheet('Unlock your Pass', (close) => {
    const inp = h('input', { type: 'password', placeholder: 'Passphrase', autocomplete: 'current-password' });
    return [h('p', { class: 'soft small' }, 'Your seed is encrypted on this device.'), h('div', { class: 'mt16' }, inp),
      h('button', { class: 'btn btn-lime mt16', style: 'width:100%', onclick: () => { close(); resolve(inp.value); } }, 'Unlock')];
  }));
  const getSeed = () => W.seed(askPass);

  function sheet(title, build) {
    const wrap = h('div', { class: 'sheet', role: 'dialog', 'aria-modal': 'true' });
    const close = () => wrap.remove();
    const box = h('div', {}, h('div', { class: 'row between' }, h('h2', { style: 'font-size:20px' }, title), h('button', { class: 'btn btn-ghost btn-sm', onclick: close, 'aria-label': 'Close' }, '✕')), ...build(close));
    wrap.append(box);
    wrap.addEventListener('mousedown', (e) => { if (e.target === wrap) close(); });
    document.body.append(wrap);
    return close;
  }

  // ------------------------------------------------------------ boot
  async function boot() {
    const vouchFor = (location.hash.match(/vouch=([0-9a-f]{64})/) || [])[1];
    const pass = W.load();
    if (vouchFor) return viewVouch(vouchFor, pass);
    if (!pass) return viewWelcome();
    if (!pass.enrollment || pass.enrollment.status !== 'active') {
      try {
        const e = await W.api('/api/enroll/' + pass.pubkey);
        pass.enrollment = e; W.save(pass);
      } catch { /* not enrolled yet */ }
    }
    if (!pass.enrollment) return viewEnroll(pass);
    return viewHome(pass);
  }

  // ------------------------------------------------------------ welcome
  function viewWelcome() {
    show(
      h('div', { class: 'center', style: 'padding:24px 0 8px' },
        h('span', { class: 'eyebrow' }, 'Singleton Pass'),
        h('h1', { style: 'font-size:34px;margin-top:12px' }, 'One pass. Every app. Nobody learns who you are.'),
        h('p', { class: 'soft mt16' }, 'Create a key on this device, verify once with an ID you already have, then prove you’re a real, unique person — and old enough — anywhere.')),
      h('div', { class: 'stack mt24' },
        h('button', { class: 'btn btn-lime btn-lg', onclick: async (e) => { busy(e.currentTarget, true); try { const p = await W.create(); viewEnroll(p); } catch (err) { alert(err.message); } finally { busy(e.currentTarget, false); } } }, 'Create my Pass'),
        h('button', { class: 'btn', onclick: viewRestore }, 'Restore from guardians')),
      h('div', { class: 'list mt24' },
        row('🔑', 'Your key stays here', 'Generated in your browser by the same Go code that runs our verifier, compiled to WebAssembly.'),
        row('🫥', 'A new pseudonym per app', 'Apps can’t compare notes about you. Even we can’t tell which pass made a proof.'),
        row('🗑', 'Your ID is never stored', 'We check the issuer’s signature, keep a keyed hash so nobody enrols twice, and discard the rest.')),
    );
  }
  function row(icon, title, sub, right) {
    return h('div', { class: 'li' }, h('div', { class: 'ico' }, icon), h('div', { class: 'grow' }, h('b', {}, title), h('small', {}, sub)), right || null);
  }

  function viewRestore() {
    const ta = h('textarea', { rows: 5, placeholder: 'Paste at least 3 shares, one per line\nsgl1-01-…' });
    const err = h('div', { class: 'notice bad hidden mt16' });
    show(
      h('button', { class: 'btn btn-ghost btn-sm', onclick: boot }, '← Back'),
      h('h1', { style: 'font-size:28px;margin:12px 0' }, 'Restore your Pass'),
      h('p', { class: 'soft' }, 'Ask three of your guardians for their recovery shares. Restoring brings back the same key — and the same pseudonym in every app.'),
      h('div', { class: 'mt16' }, ta), err,
      h('button', { class: 'btn btn-lime mt16', style: 'width:100%', onclick: async (e) => {
        busy(e.currentTarget, true); err.classList.add('hidden');
        try {
          const shares = ta.value.split(/\s+/).map((s) => s.trim()).filter(Boolean);
          const seedHex = await W.call('combine', JSON.stringify(shares));
          const p = await W.create(seedHex);
          try { p.enrollment = await W.api('/api/enroll/' + p.pubkey); W.save(p); } catch {}
          boot();
        } catch (ex) { err.textContent = ex.message; err.classList.remove('hidden'); } finally { busy(e.currentTarget, false); }
      } }, 'Restore'),
    );
  }

  // ------------------------------------------------------------ enrollment
  function viewEnroll(pass) {
    show(
      h('div', { class: 'pass-card' }, h('div', { class: 'pc-top' }, h('div', { class: 'pc-name' }, 'Singleton Pass'), h('span', { class: 'badge warn' }, 'not verified yet')), h('div', { class: 'pc-fp' }, pass.fingerprint), h('div', { class: 'pc-dot' })),
      h('h2', { style: 'font-size:22px;margin:26px 0 6px' }, 'Verify once'),
      h('p', { class: 'soft small' }, 'Pick whatever you already have. Every path ends the same way: a key that proves you’re one real person, with no ID kept anywhere.'),
      h('div', { class: 'choice mt16' },
        choice('📱', 'Mobile driver’s licence', 'Present from Apple, Google or a state wallet', 'IAL2', () => viewCredential(pass, 'mdl')),
        choice('🛂', 'Passport chip', 'Tap your phone to your passport (NFC)', 'IAL2', () => viewCredential(pass, 'passport')),
        choice('🇪🇺', 'EU Digital Identity Wallet', 'Share your PID over OpenID4VP', 'IAL2', () => viewCredential(pass, 'eudi')),
        choice('🏪', 'In person', 'Show any ID at a partner pharmacy or post office', 'IAL2', () => viewInPerson(pass)),
        choice('🤝', 'Vouching', 'No documents? Three verified people vouch for you', 'IAL1', () => viewVouchRequest(pass))),
      h('button', { class: 'btn btn-ghost btn-sm mt24', onclick: () => { if (confirm('Delete this unverified pass from the device?')) { W.clear(); boot(); } } }, 'Start over'),
    );
  }
  function choice(icon, title, sub, tag, onclick) {
    return h('button', { onclick }, h('span', { class: 'ico' }, icon), h('span', {}, h('b', {}, title), h('small', {}, sub)), h('span', { class: 'badge plain tagr ' + (tag === 'IAL2' ? 'good' : 'info') }, tag));
  }

  const FIRST = ['Maya', 'Leo', 'Amara', 'Kenji', 'Sofia', 'Noah', 'Priya', 'Mateo', 'Hana', 'Elias', 'Zara', 'Omar'];
  const LAST = ['Okafor', 'Lindqvist', 'Haddad', 'Tanaka', 'Rossi', 'Novak', 'Mehta', 'García', 'Kim', 'Murphy', 'Dubois', 'Adeyemi'];
  const pick = (a) => a[Math.floor(Math.random() * a.length)];
  const randDoc = (p) => p + Math.random().toString(36).slice(2, 9).toUpperCase();
  const isoYearsAgo = (y) => { const d = new Date(); d.setFullYear(d.getFullYear() - y); d.setDate(d.getDate() - Math.floor(Math.random() * 300)); return d.toISOString().slice(0, 10); };

  function viewCredential(pass, kind) {
    const issuers = ISSUERS.filter((i) => i.kind === kind);
    const title = { mdl: 'Mobile driver’s licence', passport: 'Passport chip', eudi: 'EU Digital Identity Wallet' }[kind];
    const f = {
      issuer: h('select', {}, issuers.map((i) => h('option', { value: i.id }, i.name))),
      given: h('input', { value: pick(FIRST) }), family: h('input', { value: pick(LAST) }),
      birth: h('input', { type: 'date', value: isoYearsAgo(19 + Math.floor(Math.random() * 30)) }),
      doc: h('input', { value: randDoc(kind === 'mdl' ? 'D' : kind === 'passport' ? 'P' : 'ID') }),
    };
    const ages = h('div', { class: 'row wrap mt8' },
      ...[['Teen (15)', 15], ['17', 17], ['Adult (19)', 19], ['30', 30], ['65', 65]].map(([l, y]) => h('button', { class: 'btn btn-sm', type: 'button', onclick: () => (f.birth.value = isoYearsAgo(y)) }, l)));
    const status = h('div', { class: 'progress-steps' });
    const err = h('div', { class: 'notice bad hidden' });
    show(
      h('button', { class: 'btn btn-ghost btn-sm', onclick: () => viewEnroll(pass) }, '← Other methods'),
      h('h1', { style: 'font-size:26px;margin:12px 0 6px' }, title),
      h('div', { class: 'notice warn mt8' }, 'Sandbox: this simulates the credential your wallet would present. A sandbox issuer signs it; Singleton verifies the signature exactly as it would a real issuer’s.'),
      h('div', { class: 'stack mt16' },
        h('label', { class: 'f' }, 'Issuer', f.issuer),
        h('div', { class: 'form-grid' }, h('label', { class: 'f' }, 'Given name', f.given), h('label', { class: 'f' }, 'Family name', f.family)),
        h('label', { class: 'f' }, 'Date of birth', f.birth, ages),
        h('label', { class: 'f' }, 'Document number', f.doc)),
      status, err,
      h('button', { class: 'btn btn-lime btn-lg mt16', style: 'width:100%', onclick: async (e) => {
        const btn = e.currentTarget; busy(btn, true); err.classList.add('hidden');
        const step = (t, cls = 'on') => status.append(h('div', { class: cls }, t));
        status.replaceChildren();
        try {
          step('→ wallet presents the credential');
          const cred = await W.api('/api/sandbox/credential', { issuer: f.issuer.value, claims: { given_name: f.given.value, family_name: f.family.value, birth_date: f.birth.value, document_number: f.doc.value } });
          step('→ Singleton checks the issuer signature');
          const enr = await W.api('/api/enroll/credential', { pubkey: pass.pubkey, credential: cred });
          step('✓ ID discarded · keyed hashes kept for de-duplication', 'done');
          step('✓ your key joined epoch sets: ' + enr.sets.map(setLabel).join(', '), 'done');
          pass.enrollment = enr; W.save(pass);
          setTimeout(() => (returnTo ? (location.href = returnTo) : viewHome(pass)), 900);
        } catch (ex) {
          err.textContent = ex.status === 409 ? ex.message + ' (one person, one pass)' : ex.message;
          err.classList.remove('hidden');
        } finally { busy(btn, false); }
      } }, 'Present credential'),
    );
  }

  function viewInPerson(pass) {
    const box = h('div', { class: 'stack mt16' }, h('div', { class: 'center muted' }, h('span', { class: 'spinner' }), ' Getting your code…'));
    show(
      h('button', { class: 'btn btn-ghost btn-sm', onclick: () => viewEnroll(pass) }, '← Other methods'),
      h('h1', { style: 'font-size:26px;margin:12px 0 6px' }, 'Verify in person'),
      h('p', { class: 'soft' }, 'Show this code and any government ID to a Singleton partner. They check the ID by hand; nothing is photographed or kept.'),
      box,
    );
    W.api('/api/enroll/in-person', { pubkey: pass.pubkey }).then(({ code }) => {
      const st = h('p', { class: 'center soft mt8' }, h('span', { class: 'spinner' }), ' Waiting for the partner to approve…');
      box.replaceChildren(h('div', { class: 'share-code' }, code), st,
        h('a', { class: 'btn', href: '/attester?code=' + encodeURIComponent(code), target: '_blank', rel: 'noopener' }, 'Demo: open the partner console ↗'));
      const timer = setInterval(async () => {
        try {
          const s = await W.api('/api/enroll/in-person/' + encodeURIComponent(code));
          if (s.status === 'approved') {
            clearInterval(timer);
            pass.enrollment = await W.api('/api/enroll/' + pass.pubkey); W.save(pass);
            st.replaceChildren('✓ Approved — welcome in.');
            setTimeout(() => (returnTo ? (location.href = returnTo) : viewHome(pass)), 800);
          } else if (s.status === 'rejected') { clearInterval(timer); st.replaceChildren('The partner could not verify you. You can try another method.'); }
        } catch { /* keep polling */ }
      }, 2500);
    }).catch((e) => box.replaceChildren(h('div', { class: 'notice bad' }, e.message)));
  }

  function viewVouchRequest(pass) {
    const link = `${location.origin}/pass#vouch=${pass.pubkey}`;
    const count = h('b', {}, '0');
    show(
      h('button', { class: 'btn btn-ghost btn-sm', onclick: () => viewEnroll(pass) }, '← Other methods'),
      h('h1', { style: 'font-size:26px;margin:12px 0 6px' }, 'Get vouched for'),
      h('p', { class: 'soft' }, 'Send this link to three people who already have a document-verified Pass. Their vouches are anonymous, rate-limited, and grant you "human" (not age).'),
      h('div', { class: 'secret mt16' }, h('code', {}, link), h('button', { class: 'btn btn-sm', onclick: () => navigator.clipboard?.writeText(link) }, 'Copy')),
      h('div', { class: 'list mt16' }, row('🤝', 'Vouches received', 'Need 3', h('span', {}, count, ' / 3'))),
    );
    const timer = setInterval(async () => {
      try {
        const v = await W.api('/api/vouch/' + pass.pubkey);
        count.textContent = v.vouches;
        if (v.vouches >= v.required) {
          clearInterval(timer);
          pass.enrollment = await W.api('/api/enroll/' + pass.pubkey); W.save(pass);
          viewHome(pass);
        }
      } catch {}
    }, 3000);
  }

  function viewVouch(vouchee, pass) {
    const fp = vouchee.slice(0, 4) + '-' + vouchee.slice(4, 8);
    const out = h('div', { class: 'mt16' });
    const canVouch = pass && pass.enrollment && (pass.enrollment.sets || []).includes('human_doc');
    show(
      h('h1', { style: 'font-size:26px;margin:12px 0 6px' }, 'Vouch for a new human'),
      h('p', { class: 'soft' }, `Someone with key ${fp}… asked you to vouch that they are a real, distinct person you know. Vouch only for people you have met. You can vouch for up to five people a month; your identity is not revealed to them or to us.`),
      canVouch ? h('button', { class: 'btn btn-lime btn-lg mt16', style: 'width:100%', onclick: async (e) => {
        busy(e.currentTarget, true);
        try {
          const seedHex = await getSeed();
          const { presentation } = await W.present(seedHex, pass.pubkey, 'singleton:vouch', 'vouch:' + vouchee, ['human_doc']);
          const r = await W.api('/api/vouch', { vouchee, presentation });
          out.replaceChildren(h('div', { class: 'notice good' }, `Thank you. They now have ${r.vouches} of ${r.required} vouches${r.enrolled ? ' — and are verified.' : '.'}`));
        } catch (ex) { out.replaceChildren(h('div', { class: 'notice bad' }, ex.message)); } finally { busy(e.currentTarget, false); }
      } }, 'I know this person — vouch') : h('div', { class: 'notice warn mt16' }, 'You need a document-verified Pass on this device to vouch.', h('a', { href: '/pass', class: 'btn btn-sm mt8' }, 'Open my Pass')),
      out,
    );
  }

  // ------------------------------------------------------------ home
  function viewHome(pass, tab = 'apps') {
    const e = pass.enrollment || {};
    const card = h('div', { class: 'pass-card' },
      h('div', { class: 'pc-top' }, h('div', { class: 'pc-name' }, 'Singleton Pass'), h('span', { class: 'badge ' + (e.status === 'active' ? 'good' : 'bad') }, e.status === 'active' ? `${e.assurance} · verified` : e.status || 'unknown')),
      h('div', { class: 'pc-fp' }, pass.fingerprint),
      h('div', { class: 'pc-sets' }, (e.sets || []).map((s) => h('span', {}, setLabel(s)))),
      h('div', { class: 'pc-dot' }));
    const upcoming = Object.entries(e.upcoming || {});
    const tabs = h('div', { class: 'tabs', role: 'tablist' }, ['apps', 'agents', 'recovery', 'settings'].map((t) => h('button', { class: t === tab ? 'on' : '', onclick: () => viewHome(pass, t) }, t[0].toUpperCase() + t.slice(1))));
    const body = { apps: tabApps, agents: tabAgents, recovery: tabRecovery, settings: tabSettings }[tab](pass);
    show(card,
      upcoming.length ? h('p', { class: 'tiny muted mt8' }, 'Coming up: ' + upcoming.map(([k, d]) => `${setLabel(k)} on ${d}`).join(' · ')) : null,
      tabs, body);
  }

  function tabApps(pass) {
    const items = (pass.history || []).slice().reverse();
    return h('div', {},
      items.length ? h('div', { class: 'list' }, items.map((x) => row(x.app ? x.app[0] : '•', x.app, `${(x.predicates || []).map(setLabel).join(', ')} · ${ago(x.at)}`, h('span', { class: 'hex', title: 'Your pseudonym in this app' }, pseud(x.nullifier)))))
        : h('div', { class: 'list' }, row('✨', 'No apps yet', 'Try a demo app to see a proof happen.')),
      h('div', { class: 'row wrap mt16' }, h('a', { class: 'btn btn-sm', href: '/demo/driftwood' }, 'Driftwood'), h('a', { class: 'btn btn-sm', href: '/demo/frontrow' }, 'Front Row'), h('a', { class: 'btn btn-sm', href: '/demo/commons' }, 'Commons')),
      h('p', { class: 'tiny muted mt16' }, 'Each app sees a different pseudonym. The list above is stored only on this device.'));
  }

  function tabAgents(pass) {
    const list = h('div', { class: 'list' }, (pass.agents || []).length ? pass.agents.map((a) => row('🤖', a.label, `${a.appName} · ${a.permissions.join(', ')} · until ${a.notAfter.slice(0, 10)}`,
      h('button', { class: 'btn btn-sm', onclick: () => agentDetails(a) }, 'Use'))) : [row('🤖', 'No agents yet', 'Let an AI agent act for you inside one app, with limits you choose.')]);
    return h('div', {}, list, h('button', { class: 'btn btn-lime mt16', style: 'width:100%', onclick: () => newAgent(pass) }, 'Delegate an agent'));
  }

  async function newAgent(pass) {
    let apps = [];
    try { apps = await W.api('/demo/apps.json'); } catch {}
    sheet('Delegate an agent', (close) => {
      const app = h('select', {}, apps.map((a) => h('option', { value: a.key }, a.name)));
      const label = h('input', { value: 'Scheduling assistant', maxlength: 80 });
      const perms = h('input', { value: 'events:create, messages:send' });
      const days = h('select', {}, [['1', '24 hours'], ['7', '7 days'], ['30', '30 days']].map(([v, l]) => h('option', { value: v }, l)));
      const err = h('div', { class: 'notice bad hidden mt8' });
      return [h('p', { class: 'soft small' }, 'The app will see that this agent acts for your pseudonym there — not who you are.'),
        h('div', { class: 'stack mt16' }, h('label', { class: 'f' }, 'App', app), h('label', { class: 'f' }, 'Agent name', label), h('label', { class: 'f' }, 'Permissions (comma separated)', perms), h('label', { class: 'f' }, 'Expires after', days)), err,
        h('button', { class: 'btn btn-lime mt16', style: 'width:100%', onclick: async (e) => {
          busy(e.currentTarget, true); err.classList.add('hidden');
          try {
            const a = apps.find((x) => x.key === app.value);
            const seedHex = await getSeed();
            const key = await W.call('agentKey');
            const now = Math.floor(Date.now() / 1000);
            const d = { v: 1, scope: a.scope, agent: key.public, label: label.value.trim(), permissions: perms.value.split(',').map((s) => s.trim()).filter(Boolean), issuedAt: now, notAfter: now + Number(days.value) * 86400 };
            const challenge = await W.call('delegationChallenge', JSON.stringify(d));
            const { presentation } = await W.present(seedHex, pass.pubkey, a.scope, challenge, ['human']);
            const res = await fetch('/v1/delegations', { method: 'POST', headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + a.pk }, body: JSON.stringify({ delegation: d, presentation }) });
            const rec = await res.json();
            if (!res.ok) throw new Error(rec.error);
            pass.agents = [...(pass.agents || []), { id: rec.id, appKey: a.key, appName: a.name, label: d.label, permissions: d.permissions, notAfter: rec.notAfter, agentPub: key.public, agentSeed: key.private }];
            W.save(pass); close(); viewHome(pass, 'agents');
          } catch (ex) { err.textContent = ex.message; err.classList.remove('hidden'); } finally { busy(e.currentTarget, false); }
        } }, 'Sign delegation')];
    });
  }

  function agentDetails(a) {
    sheet(a.label, () => {
      const out = h('div', { class: 'mt16' });
      const msg = h('input', { value: 'POST /events {"title":"Team standup"}' });
      return [h('dl', { class: 'kv' }, h('dt', {}, 'Delegation'), h('dd', { class: 'hex' }, a.id), h('dt', {}, 'Agent key'), h('dd', { class: 'hex' }, a.agentPub), h('dt', {}, 'App'), h('dd', {}, a.appName)),
        h('p', { class: 'soft small mt16' }, 'Give your agent the delegation id and its private key. It signs each request; the app verifies with POST /v1/agents/verify.'),
        h('label', { class: 'f mt16' }, 'Test a signed request', msg), out,
        h('button', { class: 'btn btn-lime mt16', style: 'width:100%', onclick: async () => {
          try {
            const signature = await W.call('agentSign', a.agentSeed, msg.value);
            const r = await W.api(`/demo/${a.appKey}/api/agent-verify`, { delegationId: a.id, message: msg.value, signature });
            out.replaceChildren(h('div', { class: 'notice ' + (r.valid ? 'good' : 'bad') }, r.valid ? `✓ ${a.appName} accepted: agent "${r.label}" acting for pseudonym ${pseud(r.nullifier)} with [${r.permissions.join(', ')}]` : '✕ ' + r.error));
          } catch (ex) { out.replaceChildren(h('div', { class: 'notice bad' }, ex.message)); }
        } }, 'Sign & send')];
    });
  }

  function tabRecovery(pass) {
    const g = pass.guardians;
    return h('div', { class: 'stack' },
      h('div', { class: 'list' }, row('🛟', g ? `Guardians set up (${g.threshold} of ${g.n})` : 'No guardians yet', g ? `Created ${ago(g.createdAt)}. Any ${g.threshold} can restore your exact key and pseudonyms.` : 'Split your seed between 5 people you trust; any 3 can restore it.')),
      h('button', { class: 'btn btn-lime', onclick: async (e) => {
        busy(e.currentTarget, true);
        try {
          const seedHex = await getSeed();
          const shares = await W.call('split', seedHex, 5, 3);
          pass.guardians = { n: 5, threshold: 3, createdAt: new Date().toISOString() }; W.save(pass);
          sheet('Your 5 recovery shares', () => [h('p', { class: 'soft small' }, 'Give one to each guardian. Any 3 restore your Pass; 2 reveal nothing. Shares are shown once.'),
            h('div', { class: 'stack mt16' }, shares.map((s, i) => h('div', { class: 'secret' }, h('code', {}, s), h('button', { class: 'btn btn-sm', onclick: () => navigator.clipboard?.writeText(s) }, `Copy #${i + 1}`)))),
            h('button', { class: 'btn mt16', style: 'width:100%', onclick: () => window.print() }, 'Print recovery cards')]);
        } catch (ex) { alert(ex.message); } finally { busy(e.currentTarget, false); }
      } }, g ? 'Re-issue shares' : 'Set up guardians'),
      h('div', { class: 'panel panel-body mt16' },
        h('h3', {}, 'Lost device?'),
        h('p', { class: 'soft small mt8' }, 'Revoking removes this key from the next epoch so nobody can use it. Restore from guardians on your new device to keep your pseudonyms.'),
        h('button', { class: 'btn btn-danger btn-sm mt16', onclick: async () => {
          if (!confirm('Revoke this Pass? Apps will stop accepting it from the next epoch.')) return;
          try {
            const seedHex = await getSeed();
            const signature = await W.call('schnorr', seedHex, 'singleton/revoke/v1:' + pass.pubkey);
            await W.api('/api/revoke', { pubkey: pass.pubkey, signature });
            pass.enrollment.status = 'revoked'; W.save(pass); viewHome(pass, 'recovery');
          } catch (ex) { alert(ex.message); }
        } }, 'Revoke this Pass')));
  }

  function tabSettings(pass) {
    return h('div', { class: 'stack' },
      h('div', { class: 'list' },
        row('🔐', pass.seed ? 'Not locked' : 'Locked with a passphrase', pass.seed ? 'Encrypt your seed on this device.' : 'You’ll be asked for it before proofs.',
          pass.seed ? h('button', { class: 'btn btn-sm', onclick: async () => {
            const phrase = prompt('Choose a passphrase (12+ characters)');
            if (!phrase || phrase.length < 12) return alert('Use at least 12 characters.');
            pass.locked = await W.lockSeed(pass.seed, phrase); W.setUnlocked(pass.seed); delete pass.seed; W.save(pass); viewHome(pass, 'settings');
          } }, 'Lock') : null),
        row('🧾', 'Public key', pass.pubkey.slice(0, 20) + '…', h('button', { class: 'btn btn-sm', onclick: () => navigator.clipboard?.writeText(pass.pubkey) }, 'Copy')),
        row('🌐', 'Network', 'Sandbox · Base Sepolia anchoring', h('a', { class: 'btn btn-sm', href: '/transparency' }, 'Audit'))),
      h('button', { class: 'btn btn-danger', onclick: () => { if (confirm('Remove the Pass from this browser? Without guardian shares you cannot get it back.')) { W.clear(); boot(); } } }, 'Remove from this device'));
  }

  boot().catch((e) => show(h('div', { class: 'notice bad' }, e.message)));
})();
