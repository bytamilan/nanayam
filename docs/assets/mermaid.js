/* Renders fenced ```mermaid code blocks as diagrams. Kramdown has no idea
   what "mermaid" is, so it ships the raw source as a plain
   <pre><code class="language-mermaid"> — this turns those into SVGs client
   side, and only pays for the library on pages that actually have one. */
(function () {
  var blocks = document.querySelectorAll('code.language-mermaid');
  if (!blocks.length) return;

  var dark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;

  var script = document.createElement('script');
  script.src = 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js';
  script.onload = function () {
    blocks.forEach(function (code) {
      var container = document.createElement('div');
      container.className = 'mermaid';
      container.textContent = code.textContent;
      code.parentElement.replaceWith(container);
    });

    mermaid.initialize({
      startOnLoad: false,
      securityLevel: 'strict',
      theme: dark ? 'dark' : 'default',
    });
    mermaid.run({ querySelector: '.mermaid' });
  };
  document.head.appendChild(script);
})();
