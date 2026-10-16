// Shared chrome for public pages: sticky nav state and mobile menu.
(() => {
  const nav = document.getElementById('nav');
  if (!nav) return;
  const onScroll = () => nav.classList.toggle('scrolled', window.scrollY > 8);
  onScroll();
  window.addEventListener('scroll', onScroll, { passive: true });
  document.getElementById('burger')?.addEventListener('click', () => nav.classList.toggle('open'));
  nav.querySelectorAll('.nav-links a').forEach((a) => a.addEventListener('click', () => nav.classList.remove('open')));
})();
