const APP_URL = "https://app.boxmind.link";

document.getElementById("go-btn").addEventListener("click", () => {
  chrome.tabs.create({ url: APP_URL });
  window.close();
});
