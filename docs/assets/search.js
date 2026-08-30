---
---
/* Client-side search over search.json (built by Jekyll from every rendered
   page). No server, no build-time dependency, no external library — the
   index is small enough (a few dozen pages) that a plain substring/score
   pass is instant. */
(function () {
  var SEARCH_URL = "{{ '/search.json' | relative_url }}";
  var input = document.getElementById('site-search-input');
  var results = document.getElementById('site-search-results');
  if (!input || !results) return;

  var index = null;
  var indexPromise = null;
  var activeIndex = -1;

  function loadIndex() {
    if (!indexPromise) {
      indexPromise = fetch(SEARCH_URL)
        .then(function (res) { return res.json(); })
        .then(function (data) { index = data; return data; })
        .catch(function () { index = []; return index; });
    }
    return indexPromise;
  }

  function escapeHtml(str) {
    return str.replace(/[&<>"']/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
    });
  }

  function snippet(content, query) {
    var lower = content.toLowerCase();
    var at = lower.indexOf(query.toLowerCase());
    if (at === -1) return content.slice(0, 120);
    var start = Math.max(0, at - 40);
    var end = Math.min(content.length, at + query.length + 80);
    return (start > 0 ? '…' : '') + content.slice(start, end) + (end < content.length ? '…' : '');
  }

  function highlight(text, query) {
    var escaped = escapeHtml(text);
    if (!query) return escaped;
    var re = new RegExp('(' + query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') + ')', 'ig');
    return escaped.replace(re, '<mark>$1</mark>');
  }

  function score(entry, terms) {
    var title = entry.title.toLowerCase();
    var content = entry.content.toLowerCase();
    var total = 0;
    for (var i = 0; i < terms.length; i++) {
      var t = terms[i];
      if (!t) continue;
      if (title.indexOf(t) !== -1) total += 10;
      if (content.indexOf(t) !== -1) total += 1;
      if (total === 0) return -1;
    }
    return total;
  }

  function render(query) {
    if (!query) {
      results.classList.remove('is-open');
      results.innerHTML = '';
      activeIndex = -1;
      return;
    }

    var terms = query.toLowerCase().split(/\s+/).filter(Boolean);
    var scored = index
      .map(function (entry) { return { entry: entry, score: score(entry, terms) }; })
      .filter(function (s) { return s.score > 0; })
      .sort(function (a, b) { return b.score - a.score; })
      .slice(0, 8);

    if (scored.length === 0) {
      results.innerHTML = '<p class="search-empty">No matches for "' + escapeHtml(query) + '"</p>';
      results.classList.add('is-open');
      activeIndex = -1;
      return;
    }

    results.innerHTML = scored
      .map(function (s, i) {
        var entry = s.entry;
        var langBadge = entry.lang === 'ta' ? '<span class="search-lang">தமிழ்</span>' : '';
        return (
          '<a href="' + entry.url + '" class="search-result" data-index="' + i + '">' +
          '<span class="search-result-title">' + highlight(entry.title, query) + langBadge + '</span>' +
          '<span class="search-result-snippet">' + highlight(snippet(entry.content, query), query) + '</span>' +
          '</a>'
        );
      })
      .join('');
    results.classList.add('is-open');
    activeIndex = -1;
  }

  input.addEventListener('focus', loadIndex);
  input.addEventListener('input', function () {
    var query = input.value.trim();
    if (!query) { render(''); return; }
    loadIndex().then(function () { render(query); });
  });

  input.addEventListener('keydown', function (e) {
    var items = results.querySelectorAll('.search-result');
    if (e.key === 'ArrowDown' && items.length) {
      e.preventDefault();
      activeIndex = Math.min(activeIndex + 1, items.length - 1);
    } else if (e.key === 'ArrowUp' && items.length) {
      e.preventDefault();
      activeIndex = Math.max(activeIndex - 1, 0);
    } else if (e.key === 'Enter' && activeIndex >= 0 && items[activeIndex]) {
      window.location.href = items[activeIndex].getAttribute('href');
      return;
    } else if (e.key === 'Escape') {
      render('');
      input.blur();
      return;
    } else {
      return;
    }
    items.forEach(function (el, i) { el.classList.toggle('is-active', i === activeIndex); });
  });

  document.addEventListener('click', function (e) {
    if (!results.contains(e.target) && e.target !== input) render('');
  });

  // "/" jumps to search from anywhere on the page, like most modern docs sites.
  document.addEventListener('keydown', function (e) {
    if (e.key === '/' && document.activeElement !== input &&
        !/^(INPUT|TEXTAREA)$/.test(document.activeElement.tagName)) {
      e.preventDefault();
      input.focus();
    }
  });
})();
