self.addEventListener("fetch", event => {
  event.respondWith(customHeaderRequestFetch(event));
});

let base = "";
const [CURRENT_ORIGIN_WITH_TAB, _fileName] = self.location.href.split(
  "proxy-service-worker.js"
);
const [CURRENT_ORIGIN] = CURRENT_ORIGIN_WITH_TAB.split("tab");
const getUrl = url =>
  `${CURRENT_ORIGIN_WITH_TAB}proxy/${encodeURIComponent(url)}`;

const ignoreFiles = [
  `${CURRENT_ORIGIN_WITH_TAB}static/proxy-init.html`,
  `${CURRENT_ORIGIN}static/proxy-register.js`,
  `${CURRENT_ORIGIN_WITH_TAB}proxy-service-worker.js`
];
const ignoreFilesSet = new Set(ignoreFiles);

function customHeaderRequestFetch(event) {
  try {
    const { request } = event;
    const { url } = request;

    debugger;
    if (url.includes("clearProxyBase")) {
      base = "";
      return fetch(new Request("", request));
    }

    console.log(base, "  ", url);
    let newUrl = decodeURIComponent(url);
    let newRequest;

    if (ignoreFilesSet.has(url)) {
      return fetch(request);
    } else if (newUrl.startsWith(`${CURRENT_ORIGIN_WITH_TAB}`)) {
      [_, newUrl] = newUrl.split(`${CURRENT_ORIGIN_WITH_TAB}`);
      try {
        ({ origin: base } = new URL(decodeURIComponent(newUrl)));
      } catch (e) {
        newUrl = `${base}/${newUrl}`;
      }
    } else if (newUrl.startsWith(`${CURRENT_ORIGIN}`)) {
      [_, newUrl] = newUrl.split(`${CURRENT_ORIGIN}`);
      newUrl = `${base}/${newUrl}`;
    } else {
      try {
        new URL(newUrl);
      } catch (e) {
        console.error(new Error("Error while trying to parse url " + url));
      }
    }

    newUrl = getUrl(newUrl);

    if (request.mode === "navigate") {
      newRequest = new Request(newUrl);
    } else {
      newRequest = new Request(newUrl, request);
    }

    return fetch(newRequest);
  } catch (e) {
    console.error(e);
  }
}
