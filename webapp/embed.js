(function () {
  const params = new URLSearchParams(window.location.search);
  const apiBase = (params.get("api") || "/api/").replace(/\/?$/, "/");
  const tenant = params.get("tenant") || "default";
  const locale = params.get("locale") || "en";

  const domainSelect = document.getElementById("domain");
  const messagesEl = document.getElementById("messages");
  const form = document.getElementById("form");
  const input = document.getElementById("input");
  let sessionId = null;

  function append(role, text) {
    const div = document.createElement("div");
    div.className = "msg " + role;
    div.textContent = text;
    messagesEl.appendChild(div);
    messagesEl.scrollTop = messagesEl.scrollHeight;
  }

  async function api(path, options) {
    const headers = Object.assign(
      { "Content-Type": "application/json", "X-Tenant-ID": tenant, "X-Locale": locale },
      (options && options.headers) || {}
    );
    const resp = await fetch(apiBase + path.replace(/^\//, ""), Object.assign({}, options, { headers }));
    if (!resp.ok) throw new Error(await resp.text());
    return resp.json();
  }

  async function init() {
    const domains = await api("domains");
    (domains.domains || []).forEach(function (d) {
      const opt = document.createElement("option");
      opt.value = d.id;
      opt.textContent = d.name || d.id;
      domainSelect.appendChild(opt);
    });
    const session = await api("session", { method: "POST", body: "{}" });
    sessionId = session.session_id;
  }

  form.addEventListener("submit", async function (ev) {
    ev.preventDefault();
    const text = input.value.trim();
    if (!text || !sessionId) return;
    append("user", text);
    input.value = "";
    append("bot", "…");
    try {
      const body = {
        session_id: sessionId,
        text: text,
        domain_id: domainSelect.value,
      };
      const data = await api("message", { method: "POST", body: JSON.stringify(body) });
      messagesEl.lastChild.textContent = data.answer || data.message || "(no answer)";
    } catch (err) {
      messagesEl.lastChild.textContent = "Error: " + err.message;
    }
  });

  init().catch(function (err) {
    append("bot", "Init failed: " + err.message);
  });
})();
