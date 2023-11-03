// sync-dark-mode.js
const htmlEl = document.documentElement;

const observer = new MutationObserver(() => {
  if (htmlEl.getAttribute('data-theme') === 'dark') {
    htmlEl.classList.add('dark');
  } else {
    htmlEl.classList.remove('dark');
  }
});

observer.observe(htmlEl, {
  attributes: true,
  attributeFilter: ['data-theme'],
});
