const TOKEN_KEY = "boxmind-token";
const USER_KEY = "boxmind-user";

function syncAuth(retry = 0) {
  try {
    if (!chrome.runtime?.id) return;
  } catch {
    // Extension was reloaded — refresh the site tab to reconnect.
    return;
  }

  const token = localStorage.getItem(TOKEN_KEY);
  const rawUser = localStorage.getItem(USER_KEY);
  let authUser = null;

  if (rawUser) {
    try {
      authUser = JSON.parse(rawUser);
    } catch {
      authUser = null;
    }
  }

  chrome.runtime
    .sendMessage({
      type: "SYNC_AUTH",
      token,
      user: authUser,
    })
    .catch(() => {
      if (retry < 3) {
        setTimeout(() => syncAuth(retry + 1), 400 * (retry + 1));
      }
    });
}

syncAuth();
window.addEventListener("storage", () => syncAuth());
window.addEventListener("boxmind-auth-change", () => syncAuth());
window.addEventListener("focus", () => syncAuth());
