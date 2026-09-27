// MV3 background service worker. Kept intentionally minimal: it just opens the
// options page on first install so the user can paste their API URL + token.
chrome.runtime.onInstalled.addListener((details) => {
  if (details.reason === 'install') {
    chrome.runtime.openOptionsPage()
  }
})
