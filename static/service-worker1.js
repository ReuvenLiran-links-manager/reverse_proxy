self.addEventListener("fetch", event => {
  event.respondWith(customHeaderRequestFetch(event));
});

function customHeaderRequestFetch(event) {
  try {
    // debugger
    const { request } = event;
    const { url } = request;
    // decide for yourself which values you provide to mode and credentials

    if (
      url === "http://localhost:9000/html.html" ||
      url === "http://localhost:9000/register.js" ||
      url === "http://localhost:9000/service-worker.js" ||
      event.request.mode === "navigate"
    ) {
      console.log("REQUEST", request);
      return fetch(request);
    }

    const base = encodeURIComponent("https://www.google.com");
    let newUrl = url;
    console.log("9000 ORIGINAL", newUrl);

    if (newUrl === "http://localhost:9000/https%3A%2F%2Fwww.google.com") {
      newUrl = "";
    } else if (newUrl.includes("http://localhost:9000")) {
      newUrl = newUrl.split("http://localhost:9000")[1];
    }

    try {
      new URL(newUrl);
      newUrl = `http://localhost:9000/${encodeURIComponent(newUrl)}`;
    } catch (e) {
      newUrl = `http://localhost:9000/${base}${encodeURIComponent(newUrl)}`;
    }
    const newRequest = new Request(newUrl, request);
    console.log("9000 MODIFIED", newRequest.url);

    return fetch(newRequest);
    // return fetch(event.request);
  } 
  catch (e) {
    console.error(e);
  }
}
