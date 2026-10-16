/*! Singleton JS SDK — https://singleton.example/docs#sdk
 * Singleton.verify({ publishableKey, required, action, reference }) → Promise<Result>
 * Singleton.button(element, options, onResult)
 * Result: { status: 'verified'|'duplicate', token, nullifier, predicates, duplicateOf, returning, session }
 * Always verify `token` on your server (EdDSA JWT, JWKS at /.well-known/jwks.json).
 */
(function (global) {
  var script = document.currentScript;
  var BASE = (script && script.src) ? new URL(script.src).origin : global.location.origin;

  function verify(opts) {
    opts = opts || {};
    if (!opts.publishableKey) return Promise.reject(new Error('Singleton: publishableKey is required'));
    var base = opts.baseUrl || BASE;
    // Open the window synchronously (inside the click) so popup blockers allow it.
    var w = 460, h = 720;
    var left = Math.max(0, (global.screenX || 0) + ((global.outerWidth || w) - w) / 2);
    var top = Math.max(0, (global.screenY || 0) + ((global.outerHeight || h) - h) / 2);
    var popup = global.open('about:blank', 'singleton_verify', 'width=' + w + ',height=' + h + ',left=' + left + ',top=' + top);
    return fetch(base + '/v1/sessions', {
      method: 'POST',
      headers: { 'Authorization': 'Bearer ' + opts.publishableKey, 'Content-Type': 'application/json' },
      body: JSON.stringify({ required: opts.required || [], action: opts.action || '', reference: opts.reference || '' })
    }).then(function (r) {
      return r.json().then(function (j) { if (!r.ok) throw new Error(j.error || 'Singleton: could not start verification'); return j; });
    }).then(function (session) {
      var url = session.url + '?mode=popup';
      if (!popup || popup.closed) {
        location.href = session.url;
        return new Promise(function () {});
      }
      popup.location.href = url;
      return new Promise(function (resolve, reject) {
        var done = false;
        function onMsg(e) {
          if (e.origin !== base || !e.data || e.data.source !== 'singleton' || e.data.session !== session.id) return;
          if (e.data.type === 'singleton:result') { done = true; cleanup(); resolve(e.data); }
          if (e.data.type === 'singleton:cancel') { done = true; cleanup(); reject(Object.assign(new Error('The person declined'), { code: 'declined' })); }
          if (e.data.type === 'singleton:unqualified') { done = true; cleanup(); reject(Object.assign(new Error('This person cannot prove: ' + e.data.missing.join(', ')), { code: 'unqualified', missing: e.data.missing })); }
        }
        var timer = setInterval(function () {
          if (popup.closed && !done) { cleanup(); reject(Object.assign(new Error('Verification window closed'), { code: 'closed' })); }
        }, 500);
        function cleanup() { clearInterval(timer); global.removeEventListener('message', onMsg); }
        global.addEventListener('message', onMsg);
      });
    }, function (err) { if (popup) popup.close(); throw err; });
  }

  function button(el, opts, onResult) {
    if (typeof el === 'string') el = document.querySelector(el);
    var b = document.createElement('button');
    b.type = 'button';
    b.setAttribute('style', 'display:inline-flex;align-items:center;gap:10px;height:44px;padding:0 18px;border-radius:999px;border:1px solid #30303a;background:#0b0b0d;color:#f3f3f1;font:600 14px system-ui,sans-serif;cursor:pointer');
    b.innerHTML = '<svg width="18" height="18" viewBox="0 0 32 32" aria-hidden="true"><circle cx="16" cy="16" r="13" fill="none" stroke="#fff" stroke-opacity=".4" stroke-width="2"/><circle cx="16" cy="16" r="4" fill="#d4ff4f"/></svg><span>' + (opts.label || 'Verify with Singleton') + '</span>';
    b.onclick = function () {
      b.disabled = true;
      verify(opts).then(function (r) { onResult && onResult(null, r); }, function (e) { onResult && onResult(e); }).then(function () { b.disabled = false; });
    };
    el.appendChild(b);
    return b;
  }

  global.Singleton = { verify: verify, button: button, version: '1.0.0' };
})(window);
