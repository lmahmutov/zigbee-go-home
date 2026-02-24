"use strict";

// === i18n — English / Russian translation support ===

var _i18n = {
    en: {
        // Navigation
        "nav.overview": "Overview",
        "nav.devices": "Devices",
        "nav.automations": "Automations",
        "nav.network": "Network",
        "nav.permit_join": "Permit Join",
        "nav.connected": "Connected",
        "nav.disconnected": "Disconnected",

        // Overview page
        "overview.devices": "devices",
        "overview.online": "online",
        "overview.offline": "offline",
        "overview.interviewed": "Interviewed",
        "overview.pending": "Pending",
        "overview.no_devices": "No devices paired yet",
        "overview.open_permit_join": "Open Permit Join",
        "overview.recent_events": "Recent Events",
        "overview.waiting_events": "Waiting for events...",

        // Devices page
        "devices.search_placeholder": "Search devices...",
        "devices.pair_new": "Pair New Device",
        "devices.unknown": "Unknown device",
        "devices.interviewed": "Interviewed",
        "devices.pending": "Pending",
        "devices.no_devices": "No devices paired yet. Open permit join to add devices.",
        "devices.open_permit_join": "Open Permit Join",

        // Device detail page
        "detail.status": "Device Status",
        "detail.info": "Device Information",
        "detail.battery": "Battery",
        "detail.signal": "Signal",
        "detail.temperature": "Temperature",
        "detail.contact": "Contact",
        "detail.ieee": "IEEE Address",
        "detail.short_addr": "Short Address",
        "detail.manufacturer": "Manufacturer",
        "detail.model": "Model",
        "detail.joined": "Joined",
        "detail.last_seen": "Last Seen",
        "detail.controls": "Controls",
        "detail.power": "Power",
        "detail.brightness": "Brightness",
        "detail.live_attrs": "Live Attributes",
        "detail.waiting_attrs": "Waiting for attribute reports...",
        "detail.endpoints": "Endpoints",
        "detail.endpoint": "Endpoint",
        "detail.input_clusters": "Input Clusters",
        "detail.output_clusters": "Output Clusters",
        "detail.none": "None",
        "detail.actions": "Actions",
        "detail.read_attribute": "Read Attribute",
        "detail.send_command": "Send Command",
        "detail.cluster": "Cluster",
        "detail.attribute": "Attribute",
        "detail.command": "Command",
        "detail.value": "Value",
        "detail.updated": "Updated",
        "detail.read": "Read",
        "detail.send": "Send",
        "detail.on": "On",
        "detail.off": "Off",
        "detail.toggle": "Toggle",
        "detail.delete_device": "Delete Device",
        "detail.save": "Save",
        "detail.cancel": "Cancel",
        "detail.open": "Open",
        "detail.closed": "Closed",
        "detail.interviewed": "Interviewed",
        "detail.not_interviewed": "Not Interviewed",
        "detail.unknown": "Unknown",

        // Network page
        "network.ncp_info": "NCP Information",
        "network.ncp_type": "NCP Type",
        "network.serial_port": "Serial Port",
        "network.baud_rate": "Baud Rate",
        "network.coordinator_ieee": "Coordinator IEEE",
        "network.fw_version": "Firmware Version",
        "network.stack_version": "Stack Version",
        "network.protocol_version": "Protocol Version",
        "network.net_info": "Network Information",
        "network.channel": "Channel",
        "network.pan_id": "PAN ID",
        "network.ext_pan_id": "Extended PAN ID",
        "network.devices": "Devices",
        "network.permit_join": "Permit Join",
        "network.permit_join_desc": "Allow new devices to join the network for a specified duration.",
        "network.60s": "60 seconds",
        "network.120s": "120 seconds",
        "network.254s": "254 seconds",
        "network.close": "Close",
        "network.seconds_remaining": "seconds remaining",

        // Automations page
        "auto.title": "Automations",
        "auto.new": "New Automation",
        "auto.enabled": "Enabled",
        "auto.disabled": "Disabled",
        "auto.edit": "Edit",
        "auto.delete": "Delete",
        "auto.enable": "Enable",
        "auto.disable": "Disable",
        "auto.no_automations": "No automations yet",
        "auto.create_first": "Create your first automation",
        "auto.back": "Back",
        "auto.name_placeholder": "Automation name",
        "auto.desc_placeholder": "Description (optional)",
        "auto.view_lua": "View Lua",
        "auto.save": "Save",
        "auto.run": "Run",
        "auto.generated_lua": "Generated Lua",
        "auto.close": "Close",

        // Page titles
        "page.overview": "Overview",
        "page.devices": "Devices",
        "page.network": "Network",
        "page.automations": "Automations",

        // Toast / dynamic messages (app.js)
        "toast.device_joined": "Device joined: ${v}",
        "toast.device_left": "Device left: ${v}",
        "toast.device_announce": "Device announce: ${v}",
        "toast.permit_join_updated": "Permit join updated",
        "toast.contact": "Contact: ${v}",
        "toast.contact_open": "open",
        "toast.contact_closed": "closed",
        "toast.low_battery": "Low battery: ${v}%",
        "toast.command_failed": "Command failed: ${v}",
        "toast.level_failed": "Level command failed: ${v}",
        "toast.permit_join": "Permit join: ${v}s",
        "toast.permit_join_failed": "Permit join failed: ${v}",
        "toast.device_renamed": "Device renamed",
        "toast.rename_failed": "Rename failed: ${v}",
        "toast.delete_confirm": "Delete device ${v}?",
        "toast.delete_failed": "Delete failed: ${v}",
        "toast.no_attr_selected": "No attribute selected",
        "toast.command_sent": "Command sent",

        // State labels
        "state.on": "On",
        "state.off": "Off",
        "state.open": "Open",
        "state.closed": "Closed",

        // Relative time
        "time.just_now": "just now",
        "time.seconds_ago": "${v}s ago",
        "time.minutes_ago": "${v}m ago",
        "time.hours_ago": "${v}h ago",
        "time.days_ago": "${v}d ago",

        // Action forms
        "action.endpoint": "Endpoint ${v}",

        // Automation toasts
        "auto.toast.enter_name": "Please enter a name",
        "auto.toast.updated": "Automation updated",
        "auto.toast.created": "Automation created",
        "auto.toast.save_failed": "Save failed: ${v}",
        "auto.toast.load_failed": "Failed to load script: ${v}",
        "auto.toast.toggle_failed": "Toggle failed: ${v}",
        "auto.toast.delete_confirm": "Delete this automation?",
        "auto.toast.delete_failed": "Delete failed: ${v}",
        "auto.toast.run_ok": "Script executed (${v})",
        "auto.toast.run_failed": "Script error: ${v}",
        "auto.toast.run_request_failed": "Run failed: ${v}",
        "auto.no_devices": "(no devices)",

        // Blockly toolbox categories
        "blockly.triggers": "Triggers",
        "blockly.device_actions": "Device Actions",
        "blockly.device_values": "Device Values",
        "blockly.datetime": "Date & Time",
        "blockly.notifications": "Notifications",
        "blockly.system": "System",
        "blockly.logic": "Logic",
        "blockly.loops": "Loops",
        "blockly.math": "Math",
        "blockly.text": "Text",
        "blockly.variables": "Variables",

        // Blockly block labels
        "block.when": "When",
        "block.becomes": "becomes",
        "block.do": "do",
        "block.changes_do": "changes, do",
        "block.turn_on": "Turn on",
        "block.turn_off": "Turn off",
        "block.toggle": "Toggle",
        "block.set_brightness": "Set brightness of",
        "block.to": "to",
        "block.pct": "%",
        "block.set_color": "Set color of",
        "block.hue": "hue",
        "block.sat": "sat",
        "block.send_raw": "Send raw command to",
        "block.endpoint": "endpoint",
        "block.cluster": "cluster",
        "block.command": "command",
        "block.wait": "Wait",
        "block.seconds_then": "seconds, then",
        "block.log": "Log",
        "block.event_value": "event value",
        "block.get": "get",
        "block.property": "property",
        "block.current": "current",
        "block.time_between": "time is between",
        "block.and": "and",
        "block.h": "h",
        "block.telegram_send": "Telegram send",
        "block.exec_cmd": "exec command",
        "block.run_cmd": "run command"
    },

    ru: {
        // Navigation
        "nav.overview": "\u041E\u0431\u0437\u043E\u0440",
        "nav.devices": "\u0423\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432\u0430",
        "nav.automations": "\u0410\u0432\u0442\u043E\u043C\u0430\u0442\u0438\u0437\u0430\u0446\u0438\u0438",
        "nav.network": "\u0421\u0435\u0442\u044C",
        "nav.permit_join": "\u041F\u043E\u0434\u043A\u043B\u044E\u0447\u0435\u043D\u0438\u0435",
        "nav.connected": "\u041F\u043E\u0434\u043A\u043B\u044E\u0447\u0435\u043D\u043E",
        "nav.disconnected": "\u041E\u0442\u043A\u043B\u044E\u0447\u0435\u043D\u043E",

        // Overview page
        "overview.devices": "\u0443\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432",
        "overview.online": "\u043E\u043D\u043B\u0430\u0439\u043D",
        "overview.offline": "\u043E\u0444\u043B\u0430\u0439\u043D",
        "overview.interviewed": "\u041E\u043F\u0440\u043E\u0448\u0435\u043D\u043E",
        "overview.pending": "\u041E\u0436\u0438\u0434\u0430\u043D\u0438\u0435",
        "overview.no_devices": "\u041D\u0435\u0442 \u043F\u043E\u0434\u043A\u043B\u044E\u0447\u0451\u043D\u043D\u044B\u0445 \u0443\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432",
        "overview.open_permit_join": "\u041E\u0442\u043A\u0440\u044B\u0442\u044C \u043F\u043E\u0434\u043A\u043B\u044E\u0447\u0435\u043D\u0438\u0435",
        "overview.recent_events": "\u041F\u043E\u0441\u043B\u0435\u0434\u043D\u0438\u0435 \u0441\u043E\u0431\u044B\u0442\u0438\u044F",
        "overview.waiting_events": "\u041E\u0436\u0438\u0434\u0430\u043D\u0438\u0435 \u0441\u043E\u0431\u044B\u0442\u0438\u0439...",

        // Devices page
        "devices.search_placeholder": "\u041F\u043E\u0438\u0441\u043A \u0443\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432...",
        "devices.pair_new": "\u0414\u043E\u0431\u0430\u0432\u0438\u0442\u044C \u0443\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432\u043E",
        "devices.unknown": "\u041D\u0435\u0438\u0437\u0432\u0435\u0441\u0442\u043D\u043E\u0435 \u0443\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432\u043E",
        "devices.interviewed": "\u041E\u043F\u0440\u043E\u0448\u0435\u043D\u043E",
        "devices.pending": "\u041E\u0436\u0438\u0434\u0430\u043D\u0438\u0435",
        "devices.no_devices": "\u041D\u0435\u0442 \u043F\u043E\u0434\u043A\u043B\u044E\u0447\u0451\u043D\u043D\u044B\u0445 \u0443\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432. \u041E\u0442\u043A\u0440\u043E\u0439\u0442\u0435 \u043F\u043E\u0434\u043A\u043B\u044E\u0447\u0435\u043D\u0438\u0435 \u0434\u043B\u044F \u0434\u043E\u0431\u0430\u0432\u043B\u0435\u043D\u0438\u044F.",
        "devices.open_permit_join": "\u041E\u0442\u043A\u0440\u044B\u0442\u044C \u043F\u043E\u0434\u043A\u043B\u044E\u0447\u0435\u043D\u0438\u0435",

        // Device detail page
        "detail.status": "\u0421\u043E\u0441\u0442\u043E\u044F\u043D\u0438\u0435 \u0443\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432\u0430",
        "detail.info": "\u0418\u043D\u0444\u043E\u0440\u043C\u0430\u0446\u0438\u044F \u043E\u0431 \u0443\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432\u0435",
        "detail.battery": "\u0411\u0430\u0442\u0430\u0440\u0435\u044F",
        "detail.signal": "\u0421\u0438\u0433\u043D\u0430\u043B",
        "detail.temperature": "\u0422\u0435\u043C\u043F\u0435\u0440\u0430\u0442\u0443\u0440\u0430",
        "detail.contact": "\u041A\u043E\u043D\u0442\u0430\u043A\u0442",
        "detail.ieee": "IEEE \u0430\u0434\u0440\u0435\u0441",
        "detail.short_addr": "\u041A\u043E\u0440\u043E\u0442\u043A\u0438\u0439 \u0430\u0434\u0440\u0435\u0441",
        "detail.manufacturer": "\u041F\u0440\u043E\u0438\u0437\u0432\u043E\u0434\u0438\u0442\u0435\u043B\u044C",
        "detail.model": "\u041C\u043E\u0434\u0435\u043B\u044C",
        "detail.joined": "\u041F\u043E\u0434\u043A\u043B\u044E\u0447\u0435\u043D\u043E",
        "detail.last_seen": "\u041F\u043E\u0441\u043B\u0435\u0434\u043D\u044F\u044F \u0430\u043A\u0442\u0438\u0432\u043D\u043E\u0441\u0442\u044C",
        "detail.controls": "\u0423\u043F\u0440\u0430\u0432\u043B\u0435\u043D\u0438\u0435",
        "detail.power": "\u041F\u0438\u0442\u0430\u043D\u0438\u0435",
        "detail.brightness": "\u042F\u0440\u043A\u043E\u0441\u0442\u044C",
        "detail.live_attrs": "\u0410\u0442\u0440\u0438\u0431\u0443\u0442\u044B",
        "detail.waiting_attrs": "\u041E\u0436\u0438\u0434\u0430\u043D\u0438\u0435 \u0430\u0442\u0440\u0438\u0431\u0443\u0442\u043E\u0432...",
        "detail.endpoints": "\u041A\u043E\u043D\u0435\u0447\u043D\u044B\u0435 \u0442\u043E\u0447\u043A\u0438",
        "detail.endpoint": "\u0422\u043E\u0447\u043A\u0430",
        "detail.input_clusters": "\u0412\u0445\u043E\u0434\u043D\u044B\u0435 \u043A\u043B\u0430\u0441\u0442\u0435\u0440\u044B",
        "detail.output_clusters": "\u0412\u044B\u0445\u043E\u0434\u043D\u044B\u0435 \u043A\u043B\u0430\u0441\u0442\u0435\u0440\u044B",
        "detail.none": "\u041D\u0435\u0442",
        "detail.actions": "\u0414\u0435\u0439\u0441\u0442\u0432\u0438\u044F",
        "detail.read_attribute": "\u0427\u0442\u0435\u043D\u0438\u0435 \u0430\u0442\u0440\u0438\u0431\u0443\u0442\u0430",
        "detail.send_command": "\u041E\u0442\u043F\u0440\u0430\u0432\u043A\u0430 \u043A\u043E\u043C\u0430\u043D\u0434\u044B",
        "detail.cluster": "\u041A\u043B\u0430\u0441\u0442\u0435\u0440",
        "detail.attribute": "\u0410\u0442\u0440\u0438\u0431\u0443\u0442",
        "detail.command": "\u041A\u043E\u043C\u0430\u043D\u0434\u0430",
        "detail.value": "\u0417\u043D\u0430\u0447\u0435\u043D\u0438\u0435",
        "detail.updated": "\u041E\u0431\u043D\u043E\u0432\u043B\u0435\u043D\u043E",
        "detail.read": "\u0427\u0438\u0442\u0430\u0442\u044C",
        "detail.send": "\u041E\u0442\u043F\u0440\u0430\u0432\u0438\u0442\u044C",
        "detail.on": "\u0412\u043A\u043B",
        "detail.off": "\u0412\u044B\u043A\u043B",
        "detail.toggle": "\u041F\u0435\u0440\u0435\u043A\u043B\u044E\u0447\u0438\u0442\u044C",
        "detail.delete_device": "\u0423\u0434\u0430\u043B\u0438\u0442\u044C \u0443\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432\u043E",
        "detail.save": "\u0421\u043E\u0445\u0440\u0430\u043D\u0438\u0442\u044C",
        "detail.cancel": "\u041E\u0442\u043C\u0435\u043D\u0430",
        "detail.open": "\u041E\u0442\u043A\u0440\u044B\u0442",
        "detail.closed": "\u0417\u0430\u043A\u0440\u044B\u0442",
        "detail.interviewed": "\u041E\u043F\u0440\u043E\u0448\u0435\u043D\u043E",
        "detail.not_interviewed": "\u041D\u0435 \u043E\u043F\u0440\u043E\u0448\u0435\u043D\u043E",
        "detail.unknown": "\u041D\u0435\u0438\u0437\u0432\u0435\u0441\u0442\u043D\u043E",

        // Network page
        "network.ncp_info": "\u0418\u043D\u0444\u043E\u0440\u043C\u0430\u0446\u0438\u044F \u043E NCP",
        "network.ncp_type": "\u0422\u0438\u043F NCP",
        "network.serial_port": "\u041F\u043E\u0440\u0442",
        "network.baud_rate": "\u0421\u043A\u043E\u0440\u043E\u0441\u0442\u044C",
        "network.coordinator_ieee": "IEEE \u043A\u043E\u043E\u0440\u0434\u0438\u043D\u0430\u0442\u043E\u0440\u0430",
        "network.fw_version": "\u0412\u0435\u0440\u0441\u0438\u044F \u043F\u0440\u043E\u0448\u0438\u0432\u043A\u0438",
        "network.stack_version": "\u0412\u0435\u0440\u0441\u0438\u044F \u0441\u0442\u0435\u043A\u0430",
        "network.protocol_version": "\u0412\u0435\u0440\u0441\u0438\u044F \u043F\u0440\u043E\u0442\u043E\u043A\u043E\u043B\u0430",
        "network.net_info": "\u0418\u043D\u0444\u043E\u0440\u043C\u0430\u0446\u0438\u044F \u043E \u0441\u0435\u0442\u0438",
        "network.channel": "\u041A\u0430\u043D\u0430\u043B",
        "network.pan_id": "PAN ID",
        "network.ext_pan_id": "Extended PAN ID",
        "network.devices": "\u0423\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432\u0430",
        "network.permit_join": "\u041F\u043E\u0434\u043A\u043B\u044E\u0447\u0435\u043D\u0438\u0435",
        "network.permit_join_desc": "\u0420\u0430\u0437\u0440\u0435\u0448\u0438\u0442\u044C \u043D\u043E\u0432\u044B\u043C \u0443\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432\u0430\u043C \u043F\u043E\u0434\u043A\u043B\u044E\u0447\u0430\u0442\u044C\u0441\u044F \u043A \u0441\u0435\u0442\u0438 \u043D\u0430 \u0443\u043A\u0430\u0437\u0430\u043D\u043D\u043E\u0435 \u0432\u0440\u0435\u043C\u044F.",
        "network.60s": "60 \u0441\u0435\u043A\u0443\u043D\u0434",
        "network.120s": "120 \u0441\u0435\u043A\u0443\u043D\u0434",
        "network.254s": "254 \u0441\u0435\u043A\u0443\u043D\u0434\u044B",
        "network.close": "\u0417\u0430\u043A\u0440\u044B\u0442\u044C",
        "network.seconds_remaining": "\u0441\u0435\u043A\u0443\u043D\u0434 \u043E\u0441\u0442\u0430\u043B\u043E\u0441\u044C",

        // Automations page
        "auto.title": "\u0410\u0432\u0442\u043E\u043C\u0430\u0442\u0438\u0437\u0430\u0446\u0438\u0438",
        "auto.new": "\u041D\u043E\u0432\u0430\u044F \u0430\u0432\u0442\u043E\u043C\u0430\u0442\u0438\u0437\u0430\u0446\u0438\u044F",
        "auto.enabled": "\u0412\u043A\u043B\u044E\u0447\u0435\u043D\u043E",
        "auto.disabled": "\u041E\u0442\u043A\u043B\u044E\u0447\u0435\u043D\u043E",
        "auto.edit": "\u0420\u0435\u0434\u0430\u043A\u0442\u0438\u0440\u043E\u0432\u0430\u0442\u044C",
        "auto.delete": "\u0423\u0434\u0430\u043B\u0438\u0442\u044C",
        "auto.enable": "\u0412\u043A\u043B\u044E\u0447\u0438\u0442\u044C",
        "auto.disable": "\u041E\u0442\u043A\u043B\u044E\u0447\u0438\u0442\u044C",
        "auto.no_automations": "\u041D\u0435\u0442 \u0430\u0432\u0442\u043E\u043C\u0430\u0442\u0438\u0437\u0430\u0446\u0438\u0439",
        "auto.create_first": "\u0421\u043E\u0437\u0434\u0430\u0439\u0442\u0435 \u043F\u0435\u0440\u0432\u0443\u044E \u0430\u0432\u0442\u043E\u043C\u0430\u0442\u0438\u0437\u0430\u0446\u0438\u044E",
        "auto.back": "\u041D\u0430\u0437\u0430\u0434",
        "auto.name_placeholder": "\u041D\u0430\u0437\u0432\u0430\u043D\u0438\u0435 \u0430\u0432\u0442\u043E\u043C\u0430\u0442\u0438\u0437\u0430\u0446\u0438\u0438",
        "auto.desc_placeholder": "\u041E\u043F\u0438\u0441\u0430\u043D\u0438\u0435 (\u043D\u0435\u043E\u0431\u044F\u0437\u0430\u0442\u0435\u043B\u044C\u043D\u043E)",
        "auto.view_lua": "\u041F\u043E\u043A\u0430\u0437\u0430\u0442\u044C Lua",
        "auto.save": "\u0421\u043E\u0445\u0440\u0430\u043D\u0438\u0442\u044C",
        "auto.run": "\u0417\u0430\u043F\u0443\u0441\u043A",
        "auto.generated_lua": "\u0421\u0433\u0435\u043D\u0435\u0440\u0438\u0440\u043E\u0432\u0430\u043D\u043D\u044B\u0439 Lua",
        "auto.close": "\u0417\u0430\u043A\u0440\u044B\u0442\u044C",

        // Page titles
        "page.overview": "\u041E\u0431\u0437\u043E\u0440",
        "page.devices": "\u0423\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432\u0430",
        "page.network": "\u0421\u0435\u0442\u044C",
        "page.automations": "\u0410\u0432\u0442\u043E\u043C\u0430\u0442\u0438\u0437\u0430\u0446\u0438\u0438",

        // Toast / dynamic messages
        "toast.device_joined": "\u0423\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432\u043E \u043F\u043E\u0434\u043A\u043B\u044E\u0447\u0435\u043D\u043E: ${v}",
        "toast.device_left": "\u0423\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432\u043E \u043E\u0442\u043A\u043B\u044E\u0447\u0435\u043D\u043E: ${v}",
        "toast.device_announce": "\u041E\u0431\u044A\u044F\u0432\u043B\u0435\u043D\u0438\u0435 \u0443\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432\u0430: ${v}",
        "toast.permit_join_updated": "\u041F\u043E\u0434\u043A\u043B\u044E\u0447\u0435\u043D\u0438\u0435 \u043E\u0431\u043D\u043E\u0432\u043B\u0435\u043D\u043E",
        "toast.contact": "\u041A\u043E\u043D\u0442\u0430\u043A\u0442: ${v}",
        "toast.contact_open": "\u043E\u0442\u043A\u0440\u044B\u0442",
        "toast.contact_closed": "\u0437\u0430\u043A\u0440\u044B\u0442",
        "toast.low_battery": "\u041D\u0438\u0437\u043A\u0438\u0439 \u0437\u0430\u0440\u044F\u0434: ${v}%",
        "toast.command_failed": "\u041E\u0448\u0438\u0431\u043A\u0430 \u043A\u043E\u043C\u0430\u043D\u0434\u044B: ${v}",
        "toast.level_failed": "\u041E\u0448\u0438\u0431\u043A\u0430 \u044F\u0440\u043A\u043E\u0441\u0442\u0438: ${v}",
        "toast.permit_join": "\u041F\u043E\u0434\u043A\u043B\u044E\u0447\u0435\u043D\u0438\u0435: ${v}\u0441",
        "toast.permit_join_failed": "\u041E\u0448\u0438\u0431\u043A\u0430 \u043F\u043E\u0434\u043A\u043B\u044E\u0447\u0435\u043D\u0438\u044F: ${v}",
        "toast.device_renamed": "\u0423\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432\u043E \u043F\u0435\u0440\u0435\u0438\u043C\u0435\u043D\u043E\u0432\u0430\u043D\u043E",
        "toast.rename_failed": "\u041E\u0448\u0438\u0431\u043A\u0430 \u043F\u0435\u0440\u0435\u0438\u043C\u0435\u043D\u043E\u0432\u0430\u043D\u0438\u044F: ${v}",
        "toast.delete_confirm": "\u0423\u0434\u0430\u043B\u0438\u0442\u044C \u0443\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432\u043E ${v}?",
        "toast.delete_failed": "\u041E\u0448\u0438\u0431\u043A\u0430 \u0443\u0434\u0430\u043B\u0435\u043D\u0438\u044F: ${v}",
        "toast.no_attr_selected": "\u0410\u0442\u0440\u0438\u0431\u0443\u0442 \u043D\u0435 \u0432\u044B\u0431\u0440\u0430\u043D",
        "toast.command_sent": "\u041A\u043E\u043C\u0430\u043D\u0434\u0430 \u043E\u0442\u043F\u0440\u0430\u0432\u043B\u0435\u043D\u0430",

        // State labels
        "state.on": "\u0412\u043A\u043B",
        "state.off": "\u0412\u044B\u043A\u043B",
        "state.open": "\u041E\u0442\u043A\u0440\u044B\u0442",
        "state.closed": "\u0417\u0430\u043A\u0440\u044B\u0442",

        // Relative time
        "time.just_now": "\u0442\u043E\u043B\u044C\u043A\u043E \u0447\u0442\u043E",
        "time.seconds_ago": "${v}\u0441 \u043D\u0430\u0437\u0430\u0434",
        "time.minutes_ago": "${v}\u043C \u043D\u0430\u0437\u0430\u0434",
        "time.hours_ago": "${v}\u0447 \u043D\u0430\u0437\u0430\u0434",
        "time.days_ago": "${v}\u0434 \u043D\u0430\u0437\u0430\u0434",

        // Action forms
        "action.endpoint": "\u0422\u043E\u0447\u043A\u0430 ${v}",

        // Automation toasts
        "auto.toast.enter_name": "\u0412\u0432\u0435\u0434\u0438\u0442\u0435 \u043D\u0430\u0437\u0432\u0430\u043D\u0438\u0435",
        "auto.toast.updated": "\u0410\u0432\u0442\u043E\u043C\u0430\u0442\u0438\u0437\u0430\u0446\u0438\u044F \u043E\u0431\u043D\u043E\u0432\u043B\u0435\u043D\u0430",
        "auto.toast.created": "\u0410\u0432\u0442\u043E\u043C\u0430\u0442\u0438\u0437\u0430\u0446\u0438\u044F \u0441\u043E\u0437\u0434\u0430\u043D\u0430",
        "auto.toast.save_failed": "\u041E\u0448\u0438\u0431\u043A\u0430 \u0441\u043E\u0445\u0440\u0430\u043D\u0435\u043D\u0438\u044F: ${v}",
        "auto.toast.load_failed": "\u041E\u0448\u0438\u0431\u043A\u0430 \u0437\u0430\u0433\u0440\u0443\u0437\u043A\u0438: ${v}",
        "auto.toast.toggle_failed": "\u041E\u0448\u0438\u0431\u043A\u0430 \u043F\u0435\u0440\u0435\u043A\u043B\u044E\u0447\u0435\u043D\u0438\u044F: ${v}",
        "auto.toast.delete_confirm": "\u0423\u0434\u0430\u043B\u0438\u0442\u044C \u044D\u0442\u0443 \u0430\u0432\u0442\u043E\u043C\u0430\u0442\u0438\u0437\u0430\u0446\u0438\u044E?",
        "auto.toast.delete_failed": "\u041E\u0448\u0438\u0431\u043A\u0430 \u0443\u0434\u0430\u043B\u0435\u043D\u0438\u044F: ${v}",
        "auto.toast.run_ok": "\u0421\u043A\u0440\u0438\u043F\u0442 \u0432\u044B\u043F\u043E\u043B\u043D\u0435\u043D (${v})",
        "auto.toast.run_failed": "\u041E\u0448\u0438\u0431\u043A\u0430 \u0441\u043A\u0440\u0438\u043F\u0442\u0430: ${v}",
        "auto.toast.run_request_failed": "\u041E\u0448\u0438\u0431\u043A\u0430 \u0437\u0430\u043F\u0443\u0441\u043A\u0430: ${v}",
        "auto.no_devices": "(\u043D\u0435\u0442 \u0443\u0441\u0442\u0440\u043E\u0439\u0441\u0442\u0432)",

        // Blockly toolbox categories
        "blockly.triggers": "\u0422\u0440\u0438\u0433\u0433\u0435\u0440\u044B",
        "blockly.device_actions": "\u0414\u0435\u0439\u0441\u0442\u0432\u0438\u044F",
        "blockly.device_values": "\u0417\u043D\u0430\u0447\u0435\u043D\u0438\u044F",
        "blockly.datetime": "\u0414\u0430\u0442\u0430 \u0438 \u0432\u0440\u0435\u043C\u044F",
        "blockly.notifications": "\u0423\u0432\u0435\u0434\u043E\u043C\u043B\u0435\u043D\u0438\u044F",
        "blockly.system": "\u0421\u0438\u0441\u0442\u0435\u043C\u0430",
        "blockly.logic": "\u041B\u043E\u0433\u0438\u043A\u0430",
        "blockly.loops": "\u0426\u0438\u043A\u043B\u044B",
        "blockly.math": "\u041C\u0430\u0442\u0435\u043C\u0430\u0442\u0438\u043A\u0430",
        "blockly.text": "\u0422\u0435\u043A\u0441\u0442",
        "blockly.variables": "\u041F\u0435\u0440\u0435\u043C\u0435\u043D\u043D\u044B\u0435",

        // Blockly block labels
        "block.when": "\u041A\u043E\u0433\u0434\u0430",
        "block.becomes": "\u0441\u0442\u0430\u043D\u0435\u0442",
        "block.do": "\u0432\u044B\u043F\u043E\u043B\u043D\u0438\u0442\u044C",
        "block.changes_do": "\u0438\u0437\u043C\u0435\u043D\u0438\u0442\u0441\u044F, \u0432\u044B\u043F\u043E\u043B\u043D\u0438\u0442\u044C",
        "block.turn_on": "\u0412\u043A\u043B\u044E\u0447\u0438\u0442\u044C",
        "block.turn_off": "\u0412\u044B\u043A\u043B\u044E\u0447\u0438\u0442\u044C",
        "block.toggle": "\u041F\u0435\u0440\u0435\u043A\u043B\u044E\u0447\u0438\u0442\u044C",
        "block.set_brightness": "\u042F\u0440\u043A\u043E\u0441\u0442\u044C",
        "block.to": "\u043D\u0430",
        "block.pct": "%",
        "block.set_color": "\u0426\u0432\u0435\u0442",
        "block.hue": "\u0442\u043E\u043D",
        "block.sat": "\u043D\u0430\u0441\u044B\u0449.",
        "block.send_raw": "\u041E\u0442\u043F\u0440\u0430\u0432\u0438\u0442\u044C \u043A\u043E\u043C\u0430\u043D\u0434\u0443",
        "block.endpoint": "\u0442\u043E\u0447\u043A\u0430",
        "block.cluster": "\u043A\u043B\u0430\u0441\u0442\u0435\u0440",
        "block.command": "\u043A\u043E\u043C\u0430\u043D\u0434\u0430",
        "block.wait": "\u0416\u0434\u0430\u0442\u044C",
        "block.seconds_then": "\u0441\u0435\u043A\u0443\u043D\u0434, \u0437\u0430\u0442\u0435\u043C",
        "block.log": "\u041B\u043E\u0433",
        "block.event_value": "\u0437\u043D\u0430\u0447\u0435\u043D\u0438\u0435 \u0441\u043E\u0431\u044B\u0442\u0438\u044F",
        "block.get": "\u043F\u043E\u043B\u0443\u0447\u0438\u0442\u044C",
        "block.property": "\u0441\u0432\u043E\u0439\u0441\u0442\u0432\u043E",
        "block.current": "\u0442\u0435\u043A\u0443\u0449\u0438\u0439",
        "block.time_between": "\u0432\u0440\u0435\u043C\u044F \u043C\u0435\u0436\u0434\u0443",
        "block.and": "\u0438",
        "block.h": "\u0447",
        "block.telegram_send": "Telegram",
        "block.exec_cmd": "\u0432\u044B\u043F\u043E\u043B\u043D\u0438\u0442\u044C",
        "block.run_cmd": "\u0437\u0430\u043F\u0443\u0441\u0442\u0438\u0442\u044C"
    }
};

// === Core functions ===

function getLang() {
    return localStorage.getItem("zigbee-lang") || "en";
}

function t(key, v) {
    var lang = getLang();
    var dict = _i18n[lang] || _i18n.en;
    var str = dict[key];
    if (str === undefined) {
        str = _i18n.en[key];
    }
    if (str === undefined) {
        return key;
    }
    if (v !== undefined) {
        str = str.replace(/\$\{v\}/g, String(v));
    }
    return str;
}

function setLang(lang) {
    localStorage.setItem("zigbee-lang", lang);
    document.documentElement.lang = lang;
    applyTranslations();

    // Update toggle button text
    var btn = document.getElementById("lang-toggle");
    if (btn) btn.textContent = lang === "en" ? "RU" : "EN";

    // Translate page title
    if (typeof translatePageTitle === "function") {
        translatePageTitle();
    }

    // If on automations page with Blockly loaded, reload to re-init locale
    if (window.blocklyWorkspace) {
        location.reload();
        return;
    }
}

function applyTranslations() {
    // Translate [data-i18n] elements
    var elements = document.querySelectorAll("[data-i18n]");
    for (var i = 0; i < elements.length; i++) {
        var el = elements[i];
        var key = el.getAttribute("data-i18n");
        if (key) {
            el.textContent = t(key);
        }
    }

    // Translate [data-i18n-placeholder] elements
    var placeholders = document.querySelectorAll("[data-i18n-placeholder]");
    for (var i = 0; i < placeholders.length; i++) {
        var el = placeholders[i];
        var key = el.getAttribute("data-i18n-placeholder");
        if (key) {
            el.placeholder = t(key);
        }
    }

    // Translate page title from topbar
    var titleEl = document.querySelector(".topbar-title[data-i18n]");
    if (titleEl) {
        var key = titleEl.getAttribute("data-i18n");
        if (key) {
            titleEl.textContent = t(key);
        }
    }
}

// === Self-initialize ===
(function() {
    var lang = getLang();
    document.documentElement.lang = lang;
})();
