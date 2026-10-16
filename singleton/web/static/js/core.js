// Loads the Go WebAssembly proof core and manages the holder's pass in this browser.
// The seed never leaves the device. It can be encrypted at rest with a passphrase
// (PBKDF2-SHA256, 310k iterations → AES-256-GCM via WebCrypto).
window.SingletonWallet = (() => {
  const KEY = 'singleton:pass:v1';
  let corePromise = null;

  function loadScript(src) {
    return new Promise((resolve, reject) => {
      if (window.Go) return resolve();
      const s = document.createElement('script');
      s.src = src; s.onload = resolve; s.onerror = () => reject(new Error('could not load ' + src));
      document.head.appendChild(s);
    });
  }

  function core() {
    if (corePromise) return corePromise;
    corePromise = (async () => {
      await loadScript('/static/wasm/wasm_exec.js');
      const go = new window.Go();
      const ready = new Promise((r) => window.addEventListener('singleton-core-ready', r, { once: true }));
      const resp = await fetch('/static/wasm/wallet.wasm');
      if (!resp.ok) throw new Error('wallet core unavailable (' + resp.status + ')');
      const { instance } = await WebAssembly.instantiateStreaming(resp, go.importObject);
      go.run(instance);
      await ready;
      return window.SingletonCore;
    })();
    return corePromise;
  }

  function unwrap(r) {
    if (r && r.error) throw new Error(r.error);
    return JSON.parse(r.ok);
  }

  async function call(name, ...args) {
    const c = await core();
    return unwrap(c[name](...args));
  }

  // ---------------------------------------------------------------- storage
  function load() {
    try { return JSON.parse(localStorage.getItem(KEY) || 'null'); } catch { return null; }
  }
  function save(pass) {
    try { localStorage.setItem(KEY, JSON.stringify(pass)); } catch (e) { throw new Error('This browser blocked local storage — the Pass cannot be saved here.'); }
  }
  function clear() { try { localStorage.removeItem(KEY); } catch {} }

  const enc = new TextEncoder();
  const hex = (buf) => [...new Uint8Array(buf)].map((b) => b.toString(16).padStart(2, '0')).join('');
  const unhex = (h) => new Uint8Array(h.match(/../g).map((x) => parseInt(x, 16)));

  async function deriveKey(passphrase, salt) {
    const base = await crypto.subtle.importKey('raw', enc.encode(passphrase), 'PBKDF2', false, ['deriveKey']);
    return crypto.subtle.deriveKey({ name: 'PBKDF2', salt, iterations: 310000, hash: 'SHA-256' }, base, { name: 'AES-GCM', length: 256 }, false, ['encrypt', 'decrypt']);
  }
  async function lockSeed(seedHex, passphrase) {
    const salt = crypto.getRandomValues(new Uint8Array(16));
    const iv = crypto.getRandomValues(new Uint8Array(12));
    const key = await deriveKey(passphrase, salt);
    const ct = await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, key, enc.encode(seedHex));
    return { salt: hex(salt), iv: hex(iv), ct: hex(ct) };
  }
  async function unlockSeed(locked, passphrase) {
    const key = await deriveKey(passphrase, unhex(locked.salt));
    try {
      const pt = await crypto.subtle.decrypt({ name: 'AES-GCM', iv: unhex(locked.iv) }, key, unhex(locked.ct));
      return new TextDecoder().decode(pt);
    } catch { throw new Error('Wrong passphrase'); }
  }

  let unlockedSeed = null;
  async function seed(promptFn) {
    const p = load();
    if (!p) throw new Error('No Singleton Pass on this device');
    if (p.seed) return p.seed;
    if (unlockedSeed) return unlockedSeed;
    const phrase = await promptFn();
    unlockedSeed = await unlockSeed(p.locked, phrase);
    return unlockedSeed;
  }

  async function create(seedHex) {
    const c = await core();
    const s = seedHex || c.newSeed();
    const pk = unwrap(c.publicKey(s));
    const pass = { v: 1, seed: s, pubkey: pk.pubkey, fingerprint: pk.fingerprint, createdAt: new Date().toISOString(), enrollment: null, history: [], agents: [], guardians: null };
    save(pass);
    return pass;
  }

  async function api(path, body, method) {
    const res = await fetch(path, { method: method || (body ? 'POST' : 'GET'), headers: body ? { 'Content-Type': 'application/json' } : {}, body: body ? JSON.stringify(body) : undefined });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) { const e = new Error(data.error || `HTTP ${res.status}`); e.status = res.status; throw e; }
    return data;
  }

  // Build a presentation for `required` sets under `scope` bound to `challenge`.
  async function present(seedHex, pubkey, scope, challenge, required, onStep) {
    onStep?.('epoch');
    const ep = await api('/api/epoch');
    const epoch = ep.epoch.epoch;
    const buckets = {};
    for (const set of required) {
      onStep?.('bucket', set);
      const b = await api(`/api/sets/${epoch}/${encodeURIComponent(set)}/bucket?pub=${pubkey}`);
      buckets[set] = { bits: b.bits, prefix: b.prefix, members: b.members };
    }
    onStep?.('sign');
    await new Promise((r) => setTimeout(r, 30)); // let the UI paint before the CPU-heavy step
    const c = await core();
    const t0 = performance.now();
    const pres = unwrap(c.prove(seedHex, scope, challenge, epoch, JSON.stringify(buckets), JSON.stringify(required), ep.minRing));
    return { presentation: pres, ms: Math.round(performance.now() - t0), ringSizes: required.map((s) => buckets[s].members.length), epoch };
  }

  return { core, call, load, save, clear, create, seed, lockSeed, unlockSeed, api, present, setUnlocked: (s) => (unlockedSeed = s) };
})();
