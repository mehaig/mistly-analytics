(function () {
    var script = document.currentScript;
    if (!script) {
        console.warn("Mistly: unable to locate current script element");
        return;
    }

    var siteId = script.getAttribute("data-site-id");
    if (!siteId) {
        console.warn("Mistly: data-site-id attribute is missing");
        return;
    }

    var endpoint = script.src.replace(/tracker\.js(\?.*)?$/, "collect");

    function send() {
        var payload = JSON.stringify({
            site_id: siteId,
            url: window.location.href,
            referrer: document.referrer,
            page_title: document.title,
            screen_width: window.screen.width,
            screen_height: window.screen.height
        });

        if (navigator.sendBeacon) {
            navigator.sendBeacon(
                endpoint,
                new Blob([payload], { type: "application/json" })
            );
            return;
        }

        fetch(endpoint, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: payload,
            keepalive: true
        }).catch(function () {});
    }

    send();

    var originalPushState = history.pushState;
    history.pushState = function () {
        originalPushState.apply(this, arguments);
        send();
    };

    window.addEventListener("popstate", send);
})();
