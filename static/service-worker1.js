self.addEventListener("fetch", event => {
  event.respondWith(customHeaderRequestFetch(event));
});

let base = "";
const CURRENT_ORIGIN = self.location.origin;
const getUrl = url => `${CURRENT_ORIGIN}/${encodeURIComponent(url)}`;

const ignoreFiles = [
  "html.html",
  "register.js",
  "service-worker.js"
].map(url => `${CURRENT_ORIGIN}/${url}`);
const ignoreFilesSet = new Set(ignoreFiles);

function customHeaderRequestFetch(event) {
  try {
    const { request } = event;
    const { url } = request;
    let newUrl = decodeURIComponent(url);
    let newRequest;

    if (ignoreFilesSet.has(url)) {
      return fetch(request);

    } else if (newUrl.includes(CURRENT_ORIGIN)) {
      ([_, newUrl] = newUrl.split(getUrl("")));
      try {
        ({ origin: base } = new URL(decodeURIComponent(newUrl)));
      } catch (e) {
        newUrl = `${base}/${newUrl}`;
      }

    } else {
      try {
        new URL(newUrl);
      } catch (e) {
        console.error(new Error('Error while trying to parse url ' + url))
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
