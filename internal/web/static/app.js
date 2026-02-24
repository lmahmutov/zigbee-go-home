"use strict";

// === API key (injected via meta tag) ===
var _apiKeyMeta = document.querySelector('meta[name="api-key"]');
var _apiKey = _apiKeyMeta ? _apiKeyMeta.getAttribute("content") : "";

// === Device state cache ===
const deviceStates = new Map(); // ieee -> {onOff, level, attributes, lastSeen}

// === WebSocket connection ===
let ws = null;
let wsRetryTimeout = null;

function connectWS() {
    if (wsRetryTimeout) {
        clearTimeout(wsRetryTimeout);
        wsRetryTimeout = null;
    }

    const proto = location.protocol === "https:" ? "wss:" : "ws:";
    ws = new WebSocket(proto + "//" + location.host + "/ws");

    ws.onopen = function() {
        const dot = document.getElementById("ws-dot");
        const text = document.getElementById("ws-status-text");
        if (dot) dot.classList.add("connected");
        if (text) text.textContent = t("nav.connected");
    };

    ws.onerror = function(err) {
        console.error("ws error:", err);
    };

    ws.onclose = function() {
        const dot = document.getElementById("ws-dot");
        const text = document.getElementById("ws-status-text");
        if (dot) dot.classList.remove("connected");
        if (text) text.textContent = t("nav.disconnected");
        wsRetryTimeout = setTimeout(connectWS, 3000);
    };

    ws.onmessage = function(evt) {
        try {
            const event = JSON.parse(evt.data);
            handleEvent(event);
        } catch(e) {
            console.error("ws parse error:", e);
        }
    };
}

// === Event handling ===
function handleEvent(event) {
    // Update events log (if on overview page)
    appendEventLog(event);

    // Route event to specific handlers
    switch(event.type) {
        case "attribute_report":
            handleAttributeReport(event.data);
            break;
        case "property_update":
            handlePropertyUpdate(event.data);
            break;
        case "device_joined":
            showToast(t("toast.device_joined", event.data.ieee || "unknown"));
            break;
        case "device_left":
            showToast(t("toast.device_left", event.data.ieee || "unknown"));
            break;
        case "device_announce":
            showToast(t("toast.device_announce", event.data.ieee || "unknown"));
            break;
        case "permit_join":
            showToast(t("toast.permit_join_updated"));
            break;
    }
}

function handleAttributeReport(data) {
    if (!data || !data.ieee) return;

    const ieee = data.ieee;

    // Update state cache
    if (!deviceStates.has(ieee)) {
        deviceStates.set(ieee, { attributes: new Map(), lastSeen: Date.now() });
    }
    const state = deviceStates.get(ieee);
    state.lastSeen = Date.now();

    const attrKey = data.endpoint + ":" + data.cluster_id + ":" + data.attr_id;
    state.attributes.set(attrKey, {
        endpoint: data.endpoint,
        clusterId: data.cluster_id,
        clusterName: data.cluster_name || "",
        attrId: data.attr_id,
        attrName: data.attr_name || "",
        value: data.value,
        time: Date.now()
    });

    // On/Off cluster (0x0006), attribute 0 (OnOff)
    if (data.cluster_id === 6 && data.attr_id === 0) {
        const isOn = !!data.value;
        state.onOff = isOn;
        updateDeviceTileState(ieee, isOn);
        updateDetailOnOff(ieee, isOn);
    }

    // Level Control cluster (0x0008), attribute 0 (CurrentLevel)
    if (data.cluster_id === 8 && data.attr_id === 0) {
        const level = typeof data.value === "number" ? data.value : 0;
        state.level = level;
        updateDetailLevel(ieee, level);
    }

    // Update live attribute table on detail page
    updateAttributeTable(ieee, data);
}

function handlePropertyUpdate(data) {
    if (!data || !data.ieee) return;

    var ieee = data.ieee;
    var prop = data.property || "";
    var value = data.value;

    // Update property in attribute table on device detail page
    if (window.currentDeviceIEEE === ieee) {
        var tbody = document.getElementById("attr-table-body");
        if (tbody) {
            var placeholder = document.getElementById("attr-placeholder");
            if (placeholder) {
                placeholder.remove();
            }

            var rowId = "prop-" + prop;
            var row = document.getElementById(rowId);
            if (!row) {
                row = document.createElement("tr");
                row.id = rowId;
                tbody.prepend(row);
            }

            var valueStr = typeof value === "object" ? JSON.stringify(value) : String(value);
            var src = data.source || {};
            var decoderLabel = (src.decoder || "") + " tag=" + (src.tag || "");

            row.innerHTML =
                '<td>-</td>' +
                '<td>' + escapeHtml(decoderLabel) + '</td>' +
                '<td>' + escapeHtml(prop) + '</td>' +
                '<td class="mono">' + escapeHtml(valueStr) + '</td>' +
                '<td class="muted">' + new Date().toLocaleTimeString() + '</td>';

            row.style.background = "var(--md-primary-container)";
            setTimeout(function() { row.style.background = ""; }, 1000);
        }

        // Update device status section on detail page
        updateDeviceStatus(prop, value);
    }

    // Update battery/LQI badges on list/overview pages
    updateDeviceBadge(ieee, prop, value);

    // Toast for interesting properties
    if (prop === "contact") {
        showToast(t("toast.contact", value ? t("toast.contact_open") : t("toast.contact_closed")));
    } else if (prop === "occupancy") {
        var detected = !!value && value !== 0;
        showToast(t("toast.occupancy", detected ? t("toast.occupancy_detected") : t("toast.occupancy_clear")));
    } else if (prop === "battery" && typeof value === "number" && value <= 10) {
        showToast(t("toast.low_battery", value), true);
    }
}

function updateDeviceStatus(prop, value) {
    // Show the status section when any property arrives
    var section = document.getElementById("device-status-section");

    switch (prop) {
        case "battery":
            var item = document.getElementById("status-battery");
            if (item) item.style.display = "";
            var el = document.getElementById("status-battery-pct");
            if (el) el.textContent = value + "%";
            if (section) section.style.display = "";
            break;
        case "battery_voltage":
            var item = document.getElementById("status-battery");
            if (item) item.style.display = "";
            var el = document.getElementById("status-battery-mv");
            if (el) el.textContent = value + " mV";
            if (section) section.style.display = "";
            break;
        case "temperature":
        case "device_temperature":
            var item = document.getElementById("status-temperature");
            if (item) item.style.display = "";
            var el = document.getElementById("status-temp");
            if (el) el.textContent = value + "\u00B0C";
            if (section) section.style.display = "";
            break;
        case "contact":
            var item = document.getElementById("status-contact");
            if (item) item.style.display = "";
            var el = document.getElementById("status-contact-val");
            if (el) el.textContent = value ? t("state.open") : t("state.closed");
            if (section) section.style.display = "";
            break;
        case "occupancy":
            var item = document.getElementById("status-occupancy");
            if (item) item.style.display = "";
            var el = document.getElementById("status-occupancy-val");
            var detected = !!value && value !== 0;
            if (el) el.textContent = detected ? t("state.motion") : t("state.clear");
            // Update icon color
            var icon = document.getElementById("status-occupancy-icon");
            if (icon) {
                var svg = icon.querySelector("svg");
                if (svg) svg.setAttribute("stroke", detected ? "#ff9800" : "#4caf50");
            }
            if (section) section.style.display = "";
            break;
        case "illuminance":
            var item = document.getElementById("status-illuminance");
            if (item) item.style.display = "";
            var el = document.getElementById("status-illuminance-val");
            if (el) el.textContent = value + " lx";
            if (section) section.style.display = "";
            break;
        case "humidity":
            var item = document.getElementById("status-humidity");
            if (item) item.style.display = "";
            var el = document.getElementById("status-humidity-val");
            if (el) el.textContent = value + "%";
            if (section) section.style.display = "";
            break;
    }
}

function updateDeviceBadge(ieee, prop, value) {
    if (prop === "battery") {
        var badges = document.querySelectorAll('.badge-battery[data-ieee="' + ieee + '"]');
        badges.forEach(function(el) {
            // Update text (preserve SVG icon)
            var svg = el.querySelector("svg");
            el.textContent = "";
            if (svg) el.appendChild(svg);
            el.appendChild(document.createTextNode(" " + value + "%"));
            // Update low class
            if (value <= 10) {
                el.classList.add("low");
            } else {
                el.classList.remove("low");
            }
        });
    }

    // Update sensor state on tiles
    var stateEl = document.getElementById("state-" + ieee);
    if (stateEl && stateEl.getAttribute("data-has-onoff") === "false") {
        if (prop === "occupancy") {
            var detected = !!value && value !== 0;
            stateEl.textContent = detected ? t("state.motion") : t("state.clear");
            stateEl.className = "device-tile-state" + (detected ? " on" : "");
        } else if (prop === "temperature" || prop === "device_temperature") {
            stateEl.textContent = value + "\u00B0C";
        } else if (prop === "illuminance") {
            stateEl.textContent = value + " lx";
        } else if (prop === "humidity") {
            stateEl.textContent = value + "%";
        } else if (prop === "contact") {
            stateEl.textContent = value ? t("state.open") : t("state.closed");
            stateEl.className = "device-tile-state" + (value ? " on" : "");
        }
    }
}

// === Dashboard tile updates ===
function updateDeviceTileState(ieee, isOn) {
    const stateEl = document.getElementById("state-" + ieee);
    if (stateEl) {
        stateEl.textContent = isOn ? t("state.on") : t("state.off");
        stateEl.className = "device-tile-state" + (isOn ? " on" : "");
    }

    const toggleEl = document.getElementById("toggle-" + ieee);
    if (toggleEl) {
        toggleEl.checked = isOn;
    }

    const iconEl = document.getElementById("icon-" + ieee);
    if (iconEl) {
        if (isOn) {
            iconEl.classList.add("active");
        } else {
            iconEl.classList.remove("active");
        }
    }
}

// === Device detail page updates ===
function updateDetailOnOff(ieee, isOn) {
    if (window.currentDeviceIEEE !== ieee) return;

    const stateEl = document.getElementById("detail-onoff-state");
    if (stateEl) {
        stateEl.textContent = isOn ? t("state.on") : t("state.off");
        stateEl.className = "control-value" + (isOn ? " text-accent" : "");
    }

    const toggleEl = document.getElementById("detail-toggle");
    if (toggleEl) {
        toggleEl.checked = isOn;
    }

    const iconEl = document.getElementById("detail-icon");
    if (iconEl) {
        if (isOn) {
            iconEl.classList.add("active");
        } else {
            iconEl.classList.remove("active");
        }
    }
}

function updateDetailLevel(ieee, level) {
    if (window.currentDeviceIEEE !== ieee) return;

    const slider = document.getElementById("detail-level");
    if (slider) {
        slider.value = level;
    }

    const valueEl = document.getElementById("detail-level-value");
    if (valueEl) {
        valueEl.textContent = Math.round(level / 254 * 100) + "%";
    }
}

function updateAttributeTable(ieee, data) {
    if (window.currentDeviceIEEE !== ieee) return;

    const tbody = document.getElementById("attr-table-body");
    if (!tbody) return;

    // Remove placeholder if present
    const placeholder = document.getElementById("attr-placeholder");
    if (placeholder) {
        placeholder.remove();
    }

    const rowId = "attr-" + data.endpoint + "-" + data.cluster_id + "-" + data.attr_id;
    let row = document.getElementById(rowId);

    if (!row) {
        row = document.createElement("tr");
        row.id = rowId;
        tbody.prepend(row);
    }

    const clusterLabel = data.cluster_name
        ? data.cluster_name + " (0x" + data.cluster_id.toString(16).toUpperCase().padStart(4, "0") + ")"
        : "0x" + data.cluster_id.toString(16).toUpperCase().padStart(4, "0");

    const attrLabel = data.attr_name
        ? data.attr_name + " (0x" + data.attr_id.toString(16).toUpperCase().padStart(4, "0") + ")"
        : "0x" + data.attr_id.toString(16).toUpperCase().padStart(4, "0");

    const valueStr = typeof data.value === "object" ? JSON.stringify(data.value) : String(data.value);

    row.innerHTML =
        '<td>' + escapeHtml(String(data.endpoint)) + '</td>' +
        '<td>' + escapeHtml(clusterLabel) + '</td>' +
        '<td>' + escapeHtml(attrLabel) + '</td>' +
        '<td class="mono">' + escapeHtml(valueStr) + '</td>' +
        '<td class="muted">' + new Date().toLocaleTimeString() + '</td>';

    // Flash effect
    row.style.background = "var(--md-primary-container)";
    setTimeout(function() { row.style.background = ""; }, 1000);
}

// === Events log ===
function appendEventLog(event) {
    const log = document.getElementById("events-log");
    if (!log) return;

    // Remove placeholder
    const placeholder = document.getElementById("events-placeholder");
    if (placeholder) placeholder.remove();

    const div = document.createElement("div");
    div.className = "event";
    const time = new Date().toLocaleTimeString();
    const dataStr = event.data ? JSON.stringify(event.data) : "";
    div.innerHTML = '<span class="muted">' + time + '</span> <span class="event-type">[' + escapeHtml(event.type) + ']</span> ' + escapeHtml(dataStr);
    log.prepend(div);

    // Keep max 100 events
    while (log.children.length > 100) {
        log.removeChild(log.lastChild);
    }
}

// === API helpers ===
async function apiCall(method, path, body) {
    const opts = { method: method, headers: {} };
    if (_apiKey) {
        opts.headers["X-API-Key"] = _apiKey;
    }
    if (body !== undefined) {
        opts.headers["Content-Type"] = "application/json";
        opts.body = JSON.stringify(body);
    }
    const resp = await fetch(path, opts);
    var data;
    try {
        data = await resp.json();
    } catch(e) {
        if (!resp.ok) {
            throw new Error("request failed: " + resp.status);
        }
        throw new Error("invalid response");
    }
    if (!resp.ok) {
        throw new Error(data.error || "request failed: " + resp.status);
    }
    return data;
}

// === Device controls ===
async function toggleDevice(ieee, endpoint, turnOn) {
    const cmdId = turnOn ? 1 : 0; // On=1, Off=0
    // Optimistic UI update
    updateDeviceTileState(ieee, turnOn);
    updateDetailOnOff(ieee, turnOn);

    try {
        await apiCall("POST", "/api/devices/" + ieee + "/command", {
            endpoint: endpoint,
            cluster_id: 6,
            command_id: cmdId
        });
    } catch(e) {
        showToast(t("toast.command_failed", e.message), true);
        // Revert on failure
        updateDeviceTileState(ieee, !turnOn);
        updateDetailOnOff(ieee, !turnOn);
    }
}

async function setLevel(ieee, endpoint, level) {
    level = parseInt(level);
    const pct = Math.round(level / 254 * 100);

    const valueEl = document.getElementById("detail-level-value");
    if (valueEl) valueEl.textContent = pct + "%";

    try {
        // Move to Level command (0x0008 cluster, command 0x04 = Move to Level with On/Off)
        // Payload: level (1 byte) + transition time (2 bytes, in 1/10 seconds)
        const payload = [level, 10, 0]; // level, transition=1s
        await apiCall("POST", "/api/devices/" + ieee + "/command", {
            endpoint: endpoint,
            cluster_id: 8,
            command_id: 4,
            payload: payload
        });
    } catch(e) {
        showToast(t("toast.level_failed", e.message), true);
    }
}

// === Permit Join ===
let permitJoinInterval = null;
let permitJoinEnd = 0;

async function permitJoin(duration) {
    try {
        await apiCall("POST", "/api/network/permit-join", { duration: duration });
        showToast(t("toast.permit_join", duration));

        if (duration > 0) {
            startPermitJoinTimer(duration);
        } else {
            stopPermitJoinTimer();
        }
    } catch(e) {
        showToast(t("toast.permit_join_failed", e.message), true);
    }
}

function startPermitJoinTimer(duration) {
    permitJoinEnd = Date.now() + duration * 1000;

    // Update sidebar button
    const btn = document.getElementById("sidebar-permit-join");
    if (btn) btn.classList.add("active");

    // Show timer on network page
    const timerEl = document.getElementById("permit-join-timer");
    if (timerEl) timerEl.classList.add("active");

    if (permitJoinInterval) clearInterval(permitJoinInterval);
    permitJoinInterval = setInterval(function() {
        const remaining = Math.max(0, Math.ceil((permitJoinEnd - Date.now()) / 1000));

        const countdownEl = document.getElementById("permit-join-countdown");
        if (countdownEl) countdownEl.textContent = remaining;

        const progressEl = document.getElementById("permit-join-progress");
        if (progressEl) {
            const totalDuration = duration;
            const pct = (remaining / totalDuration * 100);
            progressEl.style.width = pct + "%";
        }

        if (remaining <= 0) {
            stopPermitJoinTimer();
        }
    }, 1000);
}

function stopPermitJoinTimer() {
    if (permitJoinInterval) {
        clearInterval(permitJoinInterval);
        permitJoinInterval = null;
    }

    const btn = document.getElementById("sidebar-permit-join");
    if (btn) btn.classList.remove("active");

    const timerEl = document.getElementById("permit-join-timer");
    if (timerEl) timerEl.classList.remove("active");
}

// === Device rename ===
function startRename() {
    var display = document.getElementById("device-name-display");
    var edit = document.getElementById("device-name-edit");
    var input = document.getElementById("device-name-input");
    if (!display || !edit || !input) return;
    display.style.display = "none";
    var pencil = display.nextElementSibling;
    if (pencil && pencil.tagName === "BUTTON") pencil.style.display = "none";
    edit.style.display = "flex";
    input.focus();
    input.select();
}

async function saveRename(ieee) {
    var input = document.getElementById("device-name-input");
    if (!input) return;
    var name = input.value.trim();
    try {
        await apiCall("PATCH", "/api/devices/" + ieee, { friendly_name: name });
        var display = document.getElementById("device-name-display");
        if (display) display.textContent = name || ieee;
        cancelRename();
        // Update page title
        var title = document.querySelector(".topbar-title");
        if (title && name) title.textContent = name;
        showToast(t("toast.device_renamed"));
    } catch(e) {
        showToast(t("toast.rename_failed", e.message), true);
    }
}

function cancelRename() {
    var display = document.getElementById("device-name-display");
    var edit = document.getElementById("device-name-edit");
    if (display) display.style.display = "";
    // show pencil button
    var pencil = display ? display.nextElementSibling : null;
    if (pencil && pencil.tagName === "BUTTON") pencil.style.display = "";
    if (edit) edit.style.display = "none";
}

// === Delete device ===
async function deleteDevice(ieee) {
    if (!confirm(t("toast.delete_confirm", ieee))) return;
    try {
        await apiCall("DELETE", "/api/devices/" + ieee);
        location.href = "/devices";
    } catch(e) {
        showToast(t("toast.delete_failed", e.message), true);
    }
}

// === Read attribute ===
async function readAttribute(evt, ieee) {
    evt.preventDefault();
    const ep = parseInt(document.getElementById("read-ep").value);
    const cluster = parseInt(document.getElementById("read-cluster").value);
    const attrId = parseInt(document.getElementById("read-attrs").value);

    if (isNaN(attrId)) {
        document.getElementById("read-result").textContent = t("toast.no_attr_selected");
        return;
    }

    try {
        const result = await apiCall("POST", "/api/devices/" + ieee + "/read", {
            endpoint: ep,
            cluster_id: cluster,
            attr_ids: [attrId]
        });
        document.getElementById("read-result").textContent = JSON.stringify(result, null, 2);
    } catch(e) {
        document.getElementById("read-result").textContent = "Error: " + e.message;
    }
}

// === Send command ===
async function sendCommand(evt, ieee) {
    evt.preventDefault();
    const ep = parseInt(document.getElementById("cmd-ep").value);
    const cluster = parseInt(document.getElementById("cmd-cluster").value);
    const cmdId = parseInt(document.getElementById("cmd-id").value);

    try {
        await apiCall("POST", "/api/devices/" + ieee + "/command", {
            endpoint: ep,
            cluster_id: cluster,
            command_id: cmdId
        });
        showToast(t("toast.command_sent"));
    } catch(e) {
        showToast(t("toast.command_failed", e.message), true);
    }
}

// === Quick command ===
async function quickCommand(ieee, endpoint, clusterId, cmdId) {
    try {
        await apiCall("POST", "/api/devices/" + ieee + "/command", {
            endpoint: endpoint,
            cluster_id: clusterId,
            command_id: cmdId
        });
        showToast(t("toast.command_sent"));
    } catch(e) {
        showToast(t("toast.command_failed", e.message), true);
    }
}

// === Sidebar toggle ===
function toggleSidebar() {
    const sidebar = document.getElementById("sidebar");
    const overlay = document.getElementById("mobile-overlay");
    const isMobile = window.innerWidth <= 768;

    if (isMobile) {
        sidebar.classList.toggle("mobile-open");
        overlay.classList.toggle("hidden");
    } else {
        sidebar.classList.toggle("collapsed");
        localStorage.setItem("sidebar-collapsed", sidebar.classList.contains("collapsed"));
    }
}

function restoreSidebarState() {
    const sidebar = document.getElementById("sidebar");
    if (!sidebar) return;

    if (window.innerWidth > 768) {
        const collapsed = localStorage.getItem("sidebar-collapsed") === "true";
        if (collapsed) sidebar.classList.add("collapsed");
    }
}

// === Active nav item ===
function setActiveNav() {
    const path = location.pathname;
    const items = document.querySelectorAll(".nav-item");
    items.forEach(function(item) {
        const page = item.getAttribute("data-page");
        item.classList.remove("active");
        if (page === "overview" && path === "/") {
            item.classList.add("active");
        } else if (page === "devices" && path.startsWith("/devices")) {
            item.classList.add("active");
        } else if (page === "automations" && path.startsWith("/automations")) {
            item.classList.add("active");
        } else if (page === "network" && path === "/network") {
            item.classList.add("active");
        }
    });
}

// === Search/filter devices ===
function filterDevices(query) {
    query = query.toLowerCase().trim();
    const container = document.getElementById("device-list");
    if (!container) return;

    const tiles = container.querySelectorAll(".device-tile");
    tiles.forEach(function(tile) {
        const name = (tile.getAttribute("data-name") || "").toLowerCase();
        const ieee = (tile.getAttribute("data-ieee") || "").toLowerCase();
        const matches = !query || name.includes(query) || ieee.includes(query);
        tile.style.display = matches ? "" : "none";
    });
}

// === Relative time ===
function relativeTime(timestamp) {
    const now = Date.now() / 1000;
    const diff = now - timestamp;

    if (diff < 10) return t("time.just_now");
    if (diff < 60) return t("time.seconds_ago", Math.floor(diff));
    if (diff < 3600) return t("time.minutes_ago", Math.floor(diff / 60));
    if (diff < 86400) return t("time.hours_ago", Math.floor(diff / 3600));
    return t("time.days_ago", Math.floor(diff / 86400));
}

function updateRelativeTimes() {
    const elements = document.querySelectorAll(".last-seen[data-time]");
    elements.forEach(function(el) {
        const ts = parseInt(el.getAttribute("data-time"));
        if (ts > 0) {
            el.textContent = relativeTime(ts);
        }
    });
}

// === Toast notifications ===
function showToast(msg, isError) {
    const container = document.getElementById("toast-container");
    if (!container) return;
    const toast = document.createElement("div");
    toast.className = "toast" + (isError ? " error" : "");
    toast.textContent = msg;
    container.appendChild(toast);
    setTimeout(function() { toast.remove(); }, 4000);
}

// === Action form cascading dropdowns ===
function initActionForms() {
    if (!window.deviceMeta || !window.deviceMeta.length) return;

    populateSelect("read-ep", window.deviceMeta.map(function(ep) {
        return { value: ep.id, label: t("action.endpoint", ep.id) };
    }));
    populateSelect("cmd-ep", window.deviceMeta.map(function(ep) {
        return { value: ep.id, label: t("action.endpoint", ep.id) };
    }));

    onReadEpChange();
    onCmdEpChange();
}

function populateSelect(id, options) {
    var sel = document.getElementById(id);
    if (!sel) return;
    sel.innerHTML = "";
    options.forEach(function(opt) {
        var o = document.createElement("option");
        o.value = opt.value;
        o.textContent = opt.label;
        sel.appendChild(o);
    });
}

function getEpMeta(epId) {
    if (!window.deviceMeta) return null;
    return window.deviceMeta.find(function(e) { return e.id === epId; });
}

function clusterLabel(cl) {
    var hex = "0x" + cl.id.toString(16).toUpperCase().padStart(4, "0");
    return cl.name ? cl.name + " (" + hex + ")" : hex;
}

function attrLabel(a) {
    var hex = "0x" + a.id.toString(16).toUpperCase().padStart(4, "0");
    return a.name ? a.name + " (" + hex + ")" : hex;
}

function cmdLabel(cmd) {
    var hex = "0x" + cmd.id.toString(16).padStart(2, "0").toUpperCase();
    return cmd.name ? cmd.name + " (" + hex + ")" : hex;
}

function onReadEpChange() {
    var ep = getEpMeta(parseInt(document.getElementById("read-ep").value));
    var clusters = (ep && ep.clusters) || [];
    populateSelect("read-cluster", clusters.map(function(cl) {
        return { value: cl.id, label: clusterLabel(cl) };
    }));
    onReadClusterChange();
}

function onReadClusterChange() {
    var ep = getEpMeta(parseInt(document.getElementById("read-ep").value));
    var clusterId = parseInt(document.getElementById("read-cluster").value);
    var cl = ep && ep.clusters ? ep.clusters.find(function(c) { return c.id === clusterId; }) : null;
    var attrs = (cl && cl.attributes) || [];
    populateSelect("read-attrs", attrs.map(function(a) {
        return { value: a.id, label: attrLabel(a) };
    }));
}

function onCmdEpChange() {
    var ep = getEpMeta(parseInt(document.getElementById("cmd-ep").value));
    var clusters = (ep && ep.clusters) || [];
    populateSelect("cmd-cluster", clusters.map(function(cl) {
        return { value: cl.id, label: clusterLabel(cl) };
    }));
    onCmdClusterChange();
}

function onCmdClusterChange() {
    var ep = getEpMeta(parseInt(document.getElementById("cmd-ep").value));
    var clusterId = parseInt(document.getElementById("cmd-cluster").value);
    var cl = ep && ep.clusters ? ep.clusters.find(function(c) { return c.id === clusterId; }) : null;
    var cmds = (cl && cl.commands) || [];
    populateSelect("cmd-id", cmds.map(function(cmd) {
        return { value: cmd.id, label: cmdLabel(cmd) };
    }));
}

// === Translate page title based on path ===
function translatePageTitle() {
    var titleEl = document.getElementById("topbar-title");
    if (!titleEl) return;
    var path = location.pathname;
    var key = "";
    if (path === "/") key = "page.overview";
    else if (path === "/devices") key = "page.devices";
    else if (path === "/network") key = "page.network";
    else if (path === "/automations" || path.startsWith("/automations")) key = "page.automations";
    // Device detail pages keep their device-name title (no translation)
    if (key) titleEl.textContent = t(key);
}

// === Init lang toggle button text ===
function initLangToggle() {
    var btn = document.getElementById("lang-toggle");
    if (btn) btn.textContent = getLang() === "en" ? "RU" : "EN";
}

// === Utility ===
var _escapeDiv = document.createElement("div");
function escapeHtml(text) {
    _escapeDiv.textContent = text;
    return _escapeDiv.innerHTML;
}

// === Init ===
document.addEventListener("DOMContentLoaded", function() {
    connectWS();
    restoreSidebarState();
    setActiveNav();
    initLangToggle();
    applyTranslations();
    translatePageTitle();
    updateRelativeTimes();

    // Initialize action form dropdowns on device detail page
    if (window.deviceMeta) {
        initActionForms();
    }

    // Update relative times every 30s
    setInterval(updateRelativeTimes, 30000);
});
