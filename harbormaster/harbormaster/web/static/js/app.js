// Harbormaster front-end behaviour: chart tooltips, CSRF header for htmx, the hero
// stack demo, estimator sliders and small workspace conveniences. No inline scripts (CSP).
(function () {
  "use strict";

  // CSRF token for every htmx request.
  document.addEventListener("htmx:configRequest", function (e) {
    var meta = document.querySelector('meta[name="csrf-token"]');
    if (meta) e.detail.headers["X-CSRF-Token"] = meta.content;
  });

  // Chart tooltips from data-tip.
  var tip = null;
  function showTip(el, x, y) {
    if (!tip) { tip = document.createElement("div"); tip.className = "tip"; tip.setAttribute("role", "status"); document.body.appendChild(tip); }
    tip.textContent = el.getAttribute("data-tip");
    var w = tip.offsetWidth, h = tip.offsetHeight;
    var left = Math.min(window.innerWidth - w - 8, x + 14), top = y - h - 12;
    if (top < 8) top = y + 16;
    tip.style.left = left + "px"; tip.style.top = top + "px"; tip.style.display = "block";
  }
  document.addEventListener("mousemove", function (e) {
    var el = e.target.closest && e.target.closest("[data-tip]");
    if (el) showTip(el, e.clientX, e.clientY); else if (tip) tip.style.display = "none";
  });
  document.addEventListener("focusin", function (e) {
    var el = e.target.closest && e.target.closest("[data-tip]");
    if (el) { var r = el.getBoundingClientRect(); showTip(el, r.left + r.width / 2, r.top); }
  });

  // Hero stack demo: switch between precomputed engine results.
  var stacks = document.getElementById("stack-data");
  if (stacks) {
    var data = JSON.parse(stacks.textContent);
    var root = document.getElementById("stackcard");
    function draw(i) {
      var s = data[i];
      root.querySelector(".product").textContent = s.product;
      root.querySelector(".when").textContent = s.when;
      root.querySelector(".stack-total .value").textContent = s.rate + "%";
      root.querySelector(".stack-total .duty").textContent = s.duty;
      var layers = root.querySelector(".layers");
      layers.innerHTML = "";
      s.layers.forEach(function (l, j) {
        var d = document.createElement("div");
        d.className = "layer l-" + l.group;
        d.style.animationDelay = (j * 90) + "ms";
        var name = document.createElement("span"); name.textContent = l.label;
        var val = document.createElement("b"); val.textContent = l.rate + "%";
        d.appendChild(name); d.appendChild(val);
        layers.appendChild(d);
      });
      root.querySelector(".note").textContent = s.note;
      root.querySelectorAll(".stack-tabs button").forEach(function (b, k) { b.setAttribute("aria-pressed", k === i ? "true" : "false"); });
    }
    root.querySelectorAll(".stack-tabs button").forEach(function (b, k) { b.addEventListener("click", function () { draw(k); }); });
    draw(0);
  }

  // Range inputs mirror their value into an <output>.
  document.addEventListener("input", function (e) {
    if (e.target.type === "range") {
      var out = document.querySelector('output[for="' + e.target.id + '"]');
      if (out) out.textContent = e.target.value + (e.target.dataset.suffix || "");
    }
  });

  // Mobile nav toggle in the workspace.
  document.addEventListener("click", function (e) {
    var t = e.target.closest("[data-toggle-nav]");
    if (t) { var side = document.querySelector(".side"); side.classList.toggle("open"); t.setAttribute("aria-expanded", side.classList.contains("open")); }
    var s = e.target.closest("[data-suggest]");
    if (s) { var q = document.getElementById("question"); q.value = s.dataset.suggest; q.form.requestSubmit(); }
    var all = e.target.closest("[data-check-all]");
    if (all) {
      var box = document.getElementById(all.dataset.checkAll);
      var boxes = box.querySelectorAll('input[type=checkbox]');
      var on = Array.prototype.some.call(boxes, function (b) { return !b.checked; });
      boxes.forEach(function (b) { b.checked = on; });
      e.preventDefault();
    }
    var pr = e.target.closest("[data-print]");
    if (pr) { window.print(); }
  });

  // Copilot: clear and scroll after a reply lands.
  document.addEventListener("htmx:afterSwap", function (e) {
    if (e.target.id === "chat") {
      var q = document.getElementById("question"); if (q) { q.value = ""; q.focus(); }
      e.target.lastElementChild && e.target.lastElementChild.scrollIntoView({ behavior: "smooth", block: "end" });
    }
  });
})();
