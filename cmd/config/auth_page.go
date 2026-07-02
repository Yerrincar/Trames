package config

const authPageHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Trames Login</title>
  <style>
    body { margin: 0; min-height: 100vh; display: grid; place-items: center; font-family: system-ui, sans-serif; background: #111827; color: #f9fafb; }
    main { width: min(92vw, 420px); padding: 2rem; border: 1px solid #374151; border-radius: 18px; background: #1f2937; box-shadow: 0 20px 60px rgb(0 0 0 / 0.35); }
    h1 { margin-top: 0; }
    form { display: grid; gap: 0.75rem; margin-top: 1rem; }
    label { display: grid; gap: 0.35rem; color: #d1d5db; }
    input { border: 1px solid #4b5563; border-radius: 10px; padding: 0.7rem 0.8rem; background: #111827; color: #f9fafb; }
    button { border: 0; border-radius: 10px; padding: 0.75rem 1rem; background: #38bdf8; color: #082f49; font-weight: 700; cursor: pointer; }
    button.secondary { background: #a7f3d0; color: #064e3b; }
    nav { display: flex; gap: 0.75rem; margin: 1rem 0; }
    a { color: #7dd3fc; }
    .hidden { display: none; }
    pre { white-space: pre-wrap; min-height: 3rem; padding: 1rem; border-radius: 10px; background: #0f172a; color: #d1d5db; }
  </style>
</head>
<body>
  <main>
    <h1>Trames</h1>
    <nav id="auth-links">
      <a href="/login" data-view="login">Login</a>
      <a href="/register" data-view="register">Register</a>
    </nav>
    <div id="logged-in" class="hidden">
      <button type="button" data-action="logout">Logout</button>
      <button type="button" class="secondary" data-action="create-task">CreateTask</button>
    </div>
    <form id="auth-form">
      <label>Username <input id="username" name="username" autocomplete="username" required></label>
      <label>Password <input id="password" name="password" type="password" autocomplete="current-password" required></label>
      <button type="submit" id="submit-button">Login</button>
    </form>
    <pre id="result">Not logged in.</pre>
  </main>
  <script>
    const form = document.querySelector('#auth-form');
    const result = document.querySelector('#result');
    const authLinks = document.querySelector('#auth-links');
    const loggedIn = document.querySelector('#logged-in');
    const submitButton = document.querySelector('#submit-button');
    let view = location.pathname === '/register' ? 'register' : 'login';

    function renderLoggedIn(username) {
      form.classList.add('hidden');
      authLinks.classList.add('hidden');
      loggedIn.classList.remove('hidden');
      result.textContent = username ? 'Logged in as ' + username : 'Logged in.';
    }

    function renderLoggedOut(nextView) {
      view = nextView || view;
      form.classList.remove('hidden');
      authLinks.classList.remove('hidden');
      loggedIn.classList.add('hidden');
      submitButton.textContent = view === 'register' ? 'Register' : 'Login';
    }

    async function sendAuth(action) {
      const body = {
        username: form.username.value,
        password: form.password.value,
      };
      const response = await fetch('/users/' + action, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      const text = await response.text();
      result.textContent = response.status + ' ' + response.statusText + '\n' + text;
      if (response.ok && action === 'login') {
        const user = JSON.parse(text);
        renderLoggedIn(user.username);
      }
      if (response.ok && action === 'register') {
        renderLoggedOut('login');
      }
    }

    form.addEventListener('submit', (event) => {
      event.preventDefault();
      sendAuth(view).catch((error) => result.textContent = error.message);
    });

    authLinks.addEventListener('click', (event) => {
      const link = event.target.closest('a[data-view]');
      if (!link) return;
      event.preventDefault();
      history.pushState({}, '', link.href);
      renderLoggedOut(link.dataset.view);
    });

    document.querySelector('[data-action="logout"]').addEventListener('click', async () => {
      await fetch('/users/logout', { method: 'POST' });
      renderLoggedOut('login');
      result.textContent = 'Logged out.';
    });

    document.querySelector('[data-action="create-task"]').addEventListener('click', () => {
      result.textContent = 'CreateTask selected.';
    });

    fetch('/users/currentUser').then(async (response) => {
      if (!response.ok) {
        renderLoggedOut(view);
        return;
      }
      const user = await response.json();
      renderLoggedIn(user.username);
    });
  </script>
</body>
</html>`
