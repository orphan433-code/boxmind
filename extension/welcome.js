const APP_URL = "http://localhost:5173";

document.getElementById("go-btn").addEventListener("click", () => {
  chrome.tabs.create({ url: APP_URL });
  window.close();
});
