const DEFAULT_API_URL = "https://api.boxmind.link/api/v1";
const APP_URL = "https://app.boxmind.link";
const STORAGE = {
  token: "boxmindToken",
  user: "boxmindUser",
  apiUrl: "boxmindApiUrl",
  pending: "boxmindPendingQueue",
};

const RETRY = {
  maxAttempts: 4,
  delaysMs: [0, 800, 1800, 3500],
};

const inFlightSaves = new Set();
let savingCount = 0;
let badgeTimer = null;

function startSavingIndicator() {
  savingCount += 1;
  void chrome.action.setBadgeBackgroundColor({ color: "#6d88ff" });
  void chrome.action.setTitle({ title: "Сохраняю..." });

  if (badgeTimer) return;

  let pulseOn = true;
  badgeTimer = setInterval(() => {
    void chrome.action.setBadgeText({ text: pulseOn ? "…" : "·" });
    pulseOn = !pulseOn;
  }, 420);
}

function stopSavingIndicator() {
  savingCount = Math.max(0, savingCount - 1);
  if (savingCount > 0) return;

  if (badgeTimer) {
    clearInterval(badgeTimer);
    badgeTimer = null;
  }

  void chrome.action.setBadgeText({ text: "" });
  void chrome.action.setTitle({ title: "Сохранить в Boxmind" });
}

chrome.runtime.onInstalled.addListener(() => {
  void updateActionMode();
  void flushPendingQueue();
});

chrome.runtime.onStartup.addListener(() => {
  void updateActionMode();
  void flushPendingQueue();
});

chrome.storage.onChanged.addListener((changes, area) => {
  if (area !== "local" || !changes[STORAGE.token]) return;
  void updateActionMode();
  if (changes[STORAGE.token].newValue) {
    void flushPendingQueue();
  }
});

chrome.runtime.onMessage.addListener((message, _sender, sendResponse) => {
  if (message.type !== "SYNC_AUTH") return;

  if (message.token) {
    chrome.storage.local.set({
      [STORAGE.token]: message.token,
      [STORAGE.user]: message.user ?? null,
    });
    void flushPendingQueue();
  } else {
    chrome.storage.local.remove([STORAGE.token, STORAGE.user]);
  }

  void updateActionMode();
  sendResponse({ ok: true });
  return true;
});

chrome.action.onClicked.addListener(async (tab) => {
  if (!tab.id || !tab.url) return;

  const { [STORAGE.token]: token } = await chrome.storage.local.get(STORAGE.token);
  if (!token) return;

  if (isRestrictedUrl(tab.url)) {
    await showToast(tab.id, "Эту страницу сохранить нельзя", "error");
    return;
  }

  const normalizedUrl = normalizeUrl(tab.url);
  if (!normalizedUrl) {
    await showToast(tab.id, "Некорректная ссылка", "error");
    return;
  }

  if (inFlightSaves.has(normalizedUrl)) {
    await showToast(tab.id, "Уже сохраняю", "pending");
    return;
  }

  inFlightSaves.add(normalizedUrl);
  startSavingIndicator();
  await showToast(tab.id, "Сохраняю", "pending");
  try {
    await saveBookmarkInBackground(tab.id, token, normalizedUrl);
  } finally {
    inFlightSaves.delete(normalizedUrl);
    stopSavingIndicator();
  }
});

async function saveBookmarkInBackground(tabId, token, url) {
  try {
    await createBookmarkWithRetry(token, url);
    void flushPendingQueue();
  } catch (err) {
    const message = err instanceof Error ? err.message : "Не удалось сохранить";
    const status = err instanceof SaveError ? err.status : 0;

    if (status === 401 || message.toLowerCase().includes("unauthorized")) {
      await chrome.storage.local.remove([STORAGE.token, STORAGE.user]);
      await updateActionMode();
      await showToast(tabId, "Войди снова на сайте", "error");
      return;
    }

    if (status === 409) {
      await showToast(tabId, "Уже сохранено", "error");
      return;
    }

    await enqueuePending(url);
    await showToast(tabId, "Не сохранилось — попробую ещё", "error");
    void flushPendingQueue();
  }
}

async function flushPendingQueue() {
  const { [STORAGE.token]: token, [STORAGE.pending]: pending = [] } =
    await chrome.storage.local.get([STORAGE.token, STORAGE.pending]);

  if (!token || pending.length === 0) return;

  const remaining = [];

  for (const item of pending) {
    try {
      await createBookmarkWithRetry(token, item.url);
    } catch (err) {
      if (err instanceof SaveError && err.status === 409) {
        continue;
      }
      remaining.push(item);
    }
  }

  await chrome.storage.local.set({ [STORAGE.pending]: remaining });
}

async function enqueuePending(url) {
  const data = await chrome.storage.local.get(STORAGE.pending);
  const pending = data[STORAGE.pending] ?? [];
  if (pending.some((item) => item.url === url)) return;

  pending.push({ url, createdAt: Date.now() });
  await chrome.storage.local.set({ [STORAGE.pending]: pending });
}

async function createBookmarkWithRetry(token, url) {
  let lastError = null;

  for (let attempt = 0; attempt < RETRY.maxAttempts; attempt++) {
    if (RETRY.delaysMs[attempt] > 0) {
      await sleep(RETRY.delaysMs[attempt]);
    }

    try {
      return await createBookmark(token, url);
    } catch (err) {
      lastError = err;
      if (!isRetryable(err)) throw err;
    }
  }

  throw lastError ?? new Error("Не удалось сохранить");
}

class SaveError extends Error {
  constructor(message, status) {
    super(message);
    this.name = "SaveError";
    this.status = status;
  }
}

function isRetryable(err) {
  if (!(err instanceof SaveError)) {
    return err instanceof TypeError;
  }

  const status = err.status;
  if (status === 401 || status === 400 || status === 403 || status === 404 || status === 409) {
    return false;
  }

  return status === 0 || status === 408 || status === 429 || status >= 500;
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function normalizeUrl(raw) {
  try {
    const parsed = new URL(raw.trim());
    parsed.hash = "";
    if (parsed.pathname !== "/" && parsed.pathname.endsWith("/")) {
      parsed.pathname = parsed.pathname.slice(0, -1);
    }
    return parsed.toString();
  } catch {
    return "";
  }
}

async function updateActionMode() {
  const { [STORAGE.token]: token } = await chrome.storage.local.get(STORAGE.token);

  if (token) {
    await chrome.action.setPopup({ popup: "" });
    if (savingCount === 0) {
      await chrome.action.setTitle({ title: "Сохранить в Boxmind" });
    }
    return;
  }

  stopSavingIndicator();

  await chrome.action.setPopup({ popup: "welcome.html" });
  await chrome.action.setTitle({ title: "Boxmind — войди на сайте" });
}

updateActionMode();

function isRestrictedUrl(url) {
  return (
    url.startsWith("chrome://") ||
    url.startsWith("chrome-extension://") ||
    url.startsWith("edge://") ||
    url.startsWith("about:")
  );
}

async function getApiUrl() {
  const data = await chrome.storage.local.get(STORAGE.apiUrl);
  const value = data[STORAGE.apiUrl] ?? DEFAULT_API_URL;
  return value.trim().replace(/\/+$/, "");
}

async function createBookmark(token, url) {
  const apiUrl = await getApiUrl();

  let response;
  try {
    response = await fetch(`${apiUrl}/bookmarks`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ url }),
    });
  } catch (err) {
    throw err instanceof TypeError ? err : new SaveError("Сеть недоступна", 0);
  }

  if (!response.ok) {
    let message = `Ошибка ${response.status}`;
    try {
      const body = await response.json();
      if (body.error) message = body.error;
    } catch {
      // ignore
    }
    throw new SaveError(message, response.status);
  }

  return response.json();
}

async function showToast(tabId, message, variant) {
  try {
    await chrome.scripting.executeScript({
      target: { tabId },
      func: injectToast,
      args: [message, variant],
    });
  } catch {
    // restricted pages — ignore
  }
}

function injectToast(message, variant) {
  const existing = document.getElementById("boxmind-toast");
  if (existing) existing.remove();

  const root = document.createElement("div");
  root.id = "boxmind-toast";
  root.setAttribute("role", "status");
  root.setAttribute("aria-live", "polite");

  const isSuccess = variant === "success";
  const isPending = variant === "pending";

  let bg = "rgba(220, 38, 38, 0.96)";
  let shadow = "0 12px 40px rgba(220, 38, 38, 0.3)";
  let iconChar = "!";

  if (isSuccess) {
    bg = "rgba(22, 163, 74, 0.96)";
    shadow = "0 12px 40px rgba(22, 163, 74, 0.35)";
    iconChar = "✓";
  } else if (isPending) {
    bg = "rgba(109, 136, 255, 0.96)";
    shadow = "0 12px 40px rgba(109, 136, 255, 0.35)";
  }

  Object.assign(root.style, {
    position: "fixed",
    top: "20px",
    left: "50%",
    transform: "translateX(-50%) translateY(-12px)",
    zIndex: "2147483647",
    display: "inline-flex",
    alignItems: "center",
    gap: "10px",
    padding: "12px 18px",
    borderRadius: "999px",
    background: bg,
    color: "#fff",
    fontFamily: "system-ui, -apple-system, Segoe UI, sans-serif",
    fontSize: "14px",
    fontWeight: "600",
    letterSpacing: "0.01em",
    boxShadow: shadow,
    border: "1px solid rgba(255, 255, 255, 0.18)",
    opacity: "0",
    transition: "opacity 0.22s ease, transform 0.22s ease",
    pointerEvents: "none",
  });

  const icon = document.createElement("span");
  icon.textContent = iconChar;
  icon.style.fontSize = "15px";
  icon.style.lineHeight = "1";

  if (isPending && !document.getElementById("boxmind-toast-style")) {
    const style = document.createElement("style");
    style.id = "boxmind-toast-style";
    style.textContent =
      "@keyframes boxmind-pulse { 0%, 100% { opacity: 0.45; } 50% { opacity: 1; } }";
    document.documentElement.appendChild(style);
  }

  const text = document.createElement("span");
  text.textContent = message;

  if (isPending) {
    const line = document.createElement("span");
    line.style.display = "inline-flex";
    line.style.alignItems = "center";

    const dots = document.createElement("span");
    dots.textContent = "...";
    dots.style.display = "inline-block";
    dots.style.animation = "boxmind-pulse 0.9s ease-in-out infinite";

    line.appendChild(text);
    line.appendChild(dots);
    root.appendChild(line);
  } else {
    root.appendChild(icon);
    root.appendChild(text);
  }
  document.documentElement.appendChild(root);

  requestAnimationFrame(() => {
    root.style.opacity = "1";
    root.style.transform = "translateX(-50%) translateY(0)";
  });

  const dismissMs = isPending ? 1200 : 2400;
  window.setTimeout(() => {
    root.style.opacity = "0";
    root.style.transform = "translateX(-50%) translateY(-8px)";
    window.setTimeout(() => root.remove(), 220);
  }, dismissMs);
}
