const toggle = document.querySelector('.theme-toggle');
const savedTheme = localStorage.getItem('theme');

if (savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
  document.documentElement.dataset.theme = 'dark';
}

function updateThemeButton() {
  const dark = document.documentElement.dataset.theme === 'dark';
  toggle.setAttribute('aria-pressed', dark);
  toggle.setAttribute('aria-label', dark ? toggle.dataset.lightLabel : toggle.dataset.darkLabel);
}

updateThemeButton();

toggle.addEventListener('click', () => {
  const dark = document.documentElement.dataset.theme !== 'dark';
  document.documentElement.dataset.theme = dark ? 'dark' : 'light';
  localStorage.setItem('theme', dark ? 'dark' : 'light');
  updateThemeButton();
});

const emailCopy = document.querySelector('.email-copy');

async function copyEmail(email) {
  if (navigator.clipboard && window.isSecureContext) {
    await navigator.clipboard.writeText(email);
    return;
  }

  const textArea = document.createElement('textarea');
  textArea.value = email;
  textArea.setAttribute('readonly', '');
  textArea.style.position = 'fixed';
  textArea.style.opacity = '0';
  document.body.appendChild(textArea);
  textArea.select();
  const copied = document.execCommand('copy');
  textArea.remove();
  if (!copied) throw new Error('Copy command failed');
}

emailCopy.addEventListener('click', async () => {
  const email = emailCopy.dataset.email;
  const originalLabel = emailCopy.textContent;

  try {
    await copyEmail(email);
    emailCopy.textContent = 'Copied!';
    emailCopy.classList.add('is-copied');
    setTimeout(() => {
      emailCopy.textContent = originalLabel;
      emailCopy.classList.remove('is-copied');
    }, 1400);
  } catch (error) {
    emailCopy.textContent = 'Copy failed';
    setTimeout(() => { emailCopy.textContent = originalLabel; }, 1400);
  }
});
