// Kept service worker. Pages always come from the network and are never cached — they hold
// health data, and a shared phone must not show them after sign-out — so offline you get a
// plain offline page. Static assets are fingerprinted, so cache-first is safe for them.
const CACHE = "kept-v1"
const OFFLINE = "/offline.html"

self.addEventListener("install", (event) => {
  event.waitUntil(caches.open(CACHE).then((c) => c.addAll([OFFLINE, "/icon.svg"])).then(() => self.skipWaiting()))
})

self.addEventListener("activate", (event) => {
  event.waitUntil(caches.keys().then((keys) => Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k))))
    .then(() => self.clients.claim()))
})

self.addEventListener("fetch", (event) => {
  const req = event.request
  if (req.method !== "GET" || new URL(req.url).origin !== self.location.origin) return
  if (req.mode === "navigate") {
    event.respondWith(fetch(req).catch(() => caches.match(OFFLINE)))
  } else if (new URL(req.url).pathname.startsWith("/assets/")) {
    event.respondWith(caches.match(req).then((hit) => hit || fetch(req).then((res) => {
      const copy = res.clone(); caches.open(CACHE).then((c) => c.put(req, copy)); return res
    })))
  }
})
