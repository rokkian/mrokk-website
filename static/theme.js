(() => {
  const root = document.documentElement;
  let savedTheme;
  try { savedTheme = localStorage.getItem('theme'); } catch { /* A saved preference is optional. */ }
  root.dataset.theme = savedTheme === 'light' ? 'light' : 'dark';

  document.addEventListener('DOMContentLoaded', () => {
    const button = document.querySelector('.theme-control');
    const themeColor = document.querySelector('meta[name="theme-color"]');
    if (!button) return;
    function updateTheme() {
      const dark = root.dataset.theme === 'dark';
      button.setAttribute('aria-pressed', String(dark));
      button.setAttribute('aria-label', dark ? button.dataset.lightLabel : button.dataset.darkLabel);
      if (themeColor) themeColor.content = dark ? '#263e37' : '#e5ede2';
    }
    updateTheme();
    button.hidden = false;
    button.addEventListener('click', () => {
      root.dataset.theme = root.dataset.theme === 'dark' ? 'light' : 'dark';
      try { localStorage.setItem('theme', root.dataset.theme); } catch { /* Theme works without storage. */ }
      updateTheme();
    });
  });
})();
