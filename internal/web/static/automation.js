"use strict";

// === State ===
var blocklyWorkspace = null;
var currentScriptId = null;
var cachedDevices = null;
var _blocksAndGensDefined = false;
var _luaPreviewTimer = null;

// === Device list for dropdowns ===
async function loadDevices() {
    if (cachedDevices) return cachedDevices;
    try {
        cachedDevices = await apiCall("GET", "/api/devices");
    } catch(e) {
        cachedDevices = [];
    }
    return cachedDevices;
}

function deviceDropdown() {
    if (!cachedDevices || cachedDevices.length === 0) {
        return [[t("auto.no_devices"), ""]];
    }
    return cachedDevices.map(function(d) {
        var name = d.friendly_name
            || (d.manufacturer && d.model ? d.manufacturer + " " + d.model : "")
            || d.ieee_address;
        return [name, d.ieee_address];
    });
}

function propertyDropdown() {
    return [
        ["contact", "contact"],
        ["battery", "battery"],
        ["temperature", "temperature"],
        ["humidity", "humidity"],
        ["illuminance", "illuminance"],
        ["occupancy", "occupancy"],
        ["power", "power"],
        ["on_off", "on_off"]
    ];
}

// === Custom Blocks ===
function defineBlocks() {
    // --- Triggers ---
    Blockly.Blocks['zigbee_on_property'] = {
        init: function() {
            this.appendValueInput('VALUE')
                .appendField(t('block.when'))
                .appendField(new Blockly.FieldDropdown(deviceDropdown), 'DEVICE')
                .appendField(new Blockly.FieldDropdown(propertyDropdown), 'PROPERTY')
                .appendField(t('block.becomes'));
            this.appendStatementInput('DO')
                .appendField(t('block.do'));
            this.setColour(210);
            this.setTooltip('Trigger when a device property reaches a specific value');
        }
    };

    Blockly.Blocks['zigbee_on_any_property'] = {
        init: function() {
            this.appendStatementInput('DO')
                .appendField(t('block.when'))
                .appendField(new Blockly.FieldDropdown(deviceDropdown), 'DEVICE')
                .appendField(new Blockly.FieldDropdown(propertyDropdown), 'PROPERTY')
                .appendField(t('block.changes_do'));
            this.setColour(210);
            this.setTooltip('Trigger when a device property changes to any value');
        }
    };

    // --- Device Actions ---
    Blockly.Blocks['zigbee_turn_on'] = {
        init: function() {
            this.appendDummyInput()
                .appendField(t('block.turn_on'))
                .appendField(new Blockly.FieldDropdown(deviceDropdown), 'DEVICE');
            this.setPreviousStatement(true);
            this.setNextStatement(true);
            this.setColour(120);
            this.setTooltip('Turn on a device');
        }
    };

    Blockly.Blocks['zigbee_turn_off'] = {
        init: function() {
            this.appendDummyInput()
                .appendField(t('block.turn_off'))
                .appendField(new Blockly.FieldDropdown(deviceDropdown), 'DEVICE');
            this.setPreviousStatement(true);
            this.setNextStatement(true);
            this.setColour(120);
            this.setTooltip('Turn off a device');
        }
    };

    Blockly.Blocks['zigbee_toggle'] = {
        init: function() {
            this.appendDummyInput()
                .appendField(t('block.toggle'))
                .appendField(new Blockly.FieldDropdown(deviceDropdown), 'DEVICE');
            this.setPreviousStatement(true);
            this.setNextStatement(true);
            this.setColour(120);
            this.setTooltip('Toggle a device on/off');
        }
    };

    Blockly.Blocks['zigbee_set_brightness'] = {
        init: function() {
            this.appendValueInput('LEVEL')
                .setCheck('Number')
                .appendField(t('block.set_brightness'))
                .appendField(new Blockly.FieldDropdown(deviceDropdown), 'DEVICE')
                .appendField(t('block.to'));
            this.appendDummyInput()
                .appendField(t('block.pct'));
            this.setInputsInline(true);
            this.setPreviousStatement(true);
            this.setNextStatement(true);
            this.setColour(120);
            this.setTooltip('Set device brightness (0-100%)');
        }
    };

    Blockly.Blocks['zigbee_set_color'] = {
        init: function() {
            this.appendValueInput('HUE')
                .setCheck('Number')
                .appendField(t('block.set_color'))
                .appendField(new Blockly.FieldDropdown(deviceDropdown), 'DEVICE')
                .appendField(t('block.hue'));
            this.appendValueInput('SAT')
                .setCheck('Number')
                .appendField(t('block.sat'));
            this.setInputsInline(true);
            this.setPreviousStatement(true);
            this.setNextStatement(true);
            this.setColour(120);
            this.setTooltip('Set device color (hue 0-254, saturation 0-254)');
        }
    };

    Blockly.Blocks['zigbee_send_raw'] = {
        init: function() {
            this.appendDummyInput()
                .appendField(t('block.send_raw'))
                .appendField(new Blockly.FieldDropdown(deviceDropdown), 'DEVICE');
            this.appendValueInput('ENDPOINT')
                .setCheck('Number')
                .appendField(t('block.endpoint'));
            this.appendValueInput('CLUSTER')
                .setCheck('Number')
                .appendField(t('block.cluster'));
            this.appendValueInput('COMMAND')
                .setCheck('Number')
                .appendField(t('block.command'));
            this.setInputsInline(true);
            this.setPreviousStatement(true);
            this.setNextStatement(true);
            this.setColour(120);
            this.setTooltip('Send a raw ZCL command to a device');
        }
    };

    Blockly.Blocks['zigbee_wait'] = {
        init: function() {
            this.appendValueInput('SECONDS')
                .setCheck('Number')
                .appendField(t('block.wait'));
            this.appendStatementInput('DO')
                .appendField(t('block.seconds_then'));
            this.setInputsInline(true);
            this.setPreviousStatement(true);
            this.setNextStatement(true);
            this.setColour(120);
            this.setTooltip('Wait a number of seconds, then execute actions');
        }
    };

    Blockly.Blocks['zigbee_log'] = {
        init: function() {
            this.appendValueInput('MSG')
                .setCheck('String')
                .appendField(t('block.log'));
            this.setInputsInline(true);
            this.setPreviousStatement(true);
            this.setNextStatement(true);
            this.setColour(120);
            this.setTooltip('Log a message');
        }
    };

    // --- Values ---
    Blockly.Blocks['zigbee_event_value'] = {
        init: function() {
            this.appendDummyInput()
                .appendField(t('block.event_value'));
            this.setOutput(true);
            this.setColour(210);
            this.setTooltip('The value from the triggering event');
        }
    };

    Blockly.Blocks['zigbee_boolean'] = {
        init: function() {
            this.appendDummyInput()
                .appendField(new Blockly.FieldDropdown([
                    ['true', 'true'],
                    ['false', 'false']
                ]), 'BOOL');
            this.setOutput(true);
            this.setColour(210);
            this.setTooltip('Boolean value');
        }
    };

    Blockly.Blocks['zigbee_get_property'] = {
        init: function() {
            this.appendDummyInput()
                .appendField(t('block.get'))
                .appendField(new Blockly.FieldDropdown(deviceDropdown), 'DEVICE')
                .appendField(t('block.property'))
                .appendField(new Blockly.FieldDropdown(propertyDropdown), 'PROPERTY');
            this.setOutput(true);
            this.setColour(210);
            this.setTooltip('Get current value of a device property');
        }
    };

    // --- Date & Time ---
    Blockly.Blocks['system_datetime'] = {
        init: function() {
            this.appendDummyInput()
                .appendField(t('block.current'))
                .appendField(new Blockly.FieldDropdown([
                    ['hour', 'hour'],
                    ['minute', 'minute'],
                    ['second', 'second'],
                    ['weekday', 'weekday'],
                    ['day', 'day'],
                    ['month', 'month'],
                    ['year', 'year'],
                    ['timestamp', 'timestamp'],
                    ['time string', 'time_str'],
                    ['date string', 'date_str']
                ]), 'COMPONENT');
            this.setOutput(true);
            this.setColour(65);
            this.setTooltip('Get a component of the current date/time');
        }
    };

    Blockly.Blocks['system_time_between'] = {
        init: function() {
            this.appendDummyInput()
                .appendField(t('block.time_between'))
                .appendField(new Blockly.FieldNumber(8, 0, 23, 1), 'FROM')
                .appendField(t('block.and'))
                .appendField(new Blockly.FieldNumber(22, 0, 23, 1), 'TO')
                .appendField(t('block.h'));
            this.setOutput(true, 'Boolean');
            this.setColour(65);
            this.setTooltip('Check if current hour is between two hours (supports midnight wrapping, e.g. 22-6)');
        }
    };

    // --- Notifications ---
    Blockly.Blocks['telegram_send'] = {
        init: function() {
            this.appendValueInput('MSG')
                .setCheck('String')
                .appendField(t('block.telegram_send'));
            this.setInputsInline(true);
            this.setPreviousStatement(true);
            this.setNextStatement(true);
            this.setColour(330);
            this.setTooltip('Send a message via Telegram');
        }
    };

    Blockly.Blocks['system_log'] = {
        init: function() {
            this.appendValueInput('MSG')
                .setCheck('String')
                .appendField(t('block.log'))
                .appendField(new Blockly.FieldDropdown([
                    ['info', 'info'],
                    ['warn', 'warn'],
                    ['error', 'error'],
                    ['debug', 'debug']
                ]), 'LEVEL');
            this.setInputsInline(true);
            this.setPreviousStatement(true);
            this.setNextStatement(true);
            this.setColour(330);
            this.setTooltip('Log a message at specified level');
        }
    };

    // --- System ---
    Blockly.Blocks['system_exec'] = {
        init: function() {
            this.appendValueInput('CMD')
                .setCheck('String')
                .appendField(t('block.exec_cmd'));
            this.setInputsInline(true);
            this.setOutput(true, 'String');
            this.setColour(0);
            this.setTooltip('Execute an external command and return its output (must be in allowlist)');
        }
    };

    Blockly.Blocks['system_exec_no_output'] = {
        init: function() {
            this.appendValueInput('CMD')
                .setCheck('String')
                .appendField(t('block.run_cmd'));
            this.setInputsInline(true);
            this.setPreviousStatement(true);
            this.setNextStatement(true);
            this.setColour(0);
            this.setTooltip('Execute an external command (must be in allowlist)');
        }
    };
}

// === Lua Code Generators ===
// Blockly v11 API: forBlock functions receive (block, generator).
// Use generator.valueToCode / generator.statementToCode instead of Blockly.Lua.*.
function defineGenerators() {
    var G = Blockly.Lua;
    var ORDER_ATOMIC = 0;

    G.forBlock['zigbee_on_property'] = function(block, generator) {
        var device = block.getFieldValue('DEVICE');
        var property = block.getFieldValue('PROPERTY');
        var value = generator.valueToCode(block, 'VALUE', ORDER_ATOMIC) || 'true';
        var stmts = generator.statementToCode(block, 'DO');
        return 'zigbee.on("property_update", {ieee="' + device + '", property="' + property + '"}, function(event)\n' +
               '  if event.value == ' + value + ' then\n' + stmts + '  end\nend)\n';
    };

    G.forBlock['zigbee_on_any_property'] = function(block, generator) {
        var device = block.getFieldValue('DEVICE');
        var property = block.getFieldValue('PROPERTY');
        var stmts = generator.statementToCode(block, 'DO');
        return 'zigbee.on("property_update", {ieee="' + device + '", property="' + property + '"}, function(event)\n' +
               stmts + 'end)\n';
    };

    G.forBlock['zigbee_turn_on'] = function(block) {
        return 'zigbee.turn_on("' + block.getFieldValue('DEVICE') + '")\n';
    };

    G.forBlock['zigbee_turn_off'] = function(block) {
        return 'zigbee.turn_off("' + block.getFieldValue('DEVICE') + '")\n';
    };

    G.forBlock['zigbee_toggle'] = function(block) {
        return 'zigbee.toggle("' + block.getFieldValue('DEVICE') + '")\n';
    };

    G.forBlock['zigbee_set_brightness'] = function(block, generator) {
        var device = block.getFieldValue('DEVICE');
        var level = generator.valueToCode(block, 'LEVEL', ORDER_ATOMIC) || '100';
        return 'zigbee.set_brightness("' + device + '", math.floor(' + level + ' / 100 * 254))\n';
    };

    G.forBlock['zigbee_set_color'] = function(block, generator) {
        var device = block.getFieldValue('DEVICE');
        var hue = generator.valueToCode(block, 'HUE', ORDER_ATOMIC) || '0';
        var sat = generator.valueToCode(block, 'SAT', ORDER_ATOMIC) || '254';
        return 'zigbee.set_color("' + device + '", ' + hue + ', ' + sat + ')\n';
    };

    G.forBlock['zigbee_send_raw'] = function(block, generator) {
        var device = block.getFieldValue('DEVICE');
        var ep = generator.valueToCode(block, 'ENDPOINT', ORDER_ATOMIC) || '1';
        var cluster = generator.valueToCode(block, 'CLUSTER', ORDER_ATOMIC) || '0';
        var cmd = generator.valueToCode(block, 'COMMAND', ORDER_ATOMIC) || '0';
        return 'zigbee.send_command("' + device + '", ' + ep + ', ' + cluster + ', ' + cmd + ')\n';
    };

    G.forBlock['zigbee_wait'] = function(block, generator) {
        var seconds = generator.valueToCode(block, 'SECONDS', ORDER_ATOMIC) || '1';
        var stmts = generator.statementToCode(block, 'DO');
        return 'zigbee.after(' + seconds + ', function()\n' + stmts + 'end)\n';
    };

    G.forBlock['zigbee_log'] = function(block, generator) {
        var msg = generator.valueToCode(block, 'MSG', ORDER_ATOMIC) || '""';
        return 'zigbee.log(' + msg + ')\n';
    };

    G.forBlock['zigbee_event_value'] = function() {
        return ['event.value', ORDER_ATOMIC];
    };

    G.forBlock['zigbee_boolean'] = function(block) {
        return [block.getFieldValue('BOOL'), ORDER_ATOMIC];
    };

    G.forBlock['zigbee_get_property'] = function(block) {
        var device = block.getFieldValue('DEVICE');
        var property = block.getFieldValue('PROPERTY');
        return ['zigbee.get_property("' + device + '", "' + property + '")', ORDER_ATOMIC];
    };

    G.forBlock['system_datetime'] = function(block) {
        return ['system.datetime("' + block.getFieldValue('COMPONENT') + '")', ORDER_ATOMIC];
    };

    G.forBlock['system_time_between'] = function(block) {
        return ['system.time_between(' + block.getFieldValue('FROM') + ', ' + block.getFieldValue('TO') + ')', ORDER_ATOMIC];
    };

    G.forBlock['telegram_send'] = function(block, generator) {
        var msg = generator.valueToCode(block, 'MSG', ORDER_ATOMIC) || '""';
        return 'telegram.send(' + msg + ')\n';
    };

    G.forBlock['system_log'] = function(block, generator) {
        var level = block.getFieldValue('LEVEL');
        var msg = generator.valueToCode(block, 'MSG', ORDER_ATOMIC) || '""';
        return 'system.log("' + level + '", ' + msg + ')\n';
    };

    G.forBlock['system_exec'] = function(block, generator) {
        var cmd = generator.valueToCode(block, 'CMD', ORDER_ATOMIC) || '""';
        return ['system.exec(' + cmd + ')', ORDER_ATOMIC];
    };

    G.forBlock['system_exec_no_output'] = function(block, generator) {
        var cmd = generator.valueToCode(block, 'CMD', ORDER_ATOMIC) || '""';
        return 'system.exec(' + cmd + ')\n';
    };

    // --- Fallback generators for standard blocks that may be missing in v11 ---
    // controls_ifelse is controls_if with pre-configured else; share the generator.
    if (!G.forBlock['controls_ifelse'] && G.forBlock['controls_if']) {
        G.forBlock['controls_ifelse'] = G.forBlock['controls_if'];
    }
}

// === Toolbox ===
function buildToolbox() {
    return {
        kind: 'categoryToolbox',
        contents: [
            {
                kind: 'category',
                name: t('blockly.triggers'),
                colour: 210,
                contents: [
                    { kind: 'block', type: 'zigbee_on_property' },
                    { kind: 'block', type: 'zigbee_on_any_property' }
                ]
            },
            {
                kind: 'category',
                name: t('blockly.device_actions'),
                colour: 120,
                contents: [
                    { kind: 'block', type: 'zigbee_turn_on' },
                    { kind: 'block', type: 'zigbee_turn_off' },
                    { kind: 'block', type: 'zigbee_toggle' },
                    { kind: 'block', type: 'zigbee_set_brightness', inputs: {
                        LEVEL: { shadow: { type: 'math_number', fields: { NUM: 100 } } }
                    }},
                    { kind: 'block', type: 'zigbee_set_color', inputs: {
                        HUE: { shadow: { type: 'math_number', fields: { NUM: 0 } } },
                        SAT: { shadow: { type: 'math_number', fields: { NUM: 254 } } }
                    }},
                    { kind: 'block', type: 'zigbee_send_raw', inputs: {
                        ENDPOINT: { shadow: { type: 'math_number', fields: { NUM: 1 } } },
                        CLUSTER: { shadow: { type: 'math_number', fields: { NUM: 6 } } },
                        COMMAND: { shadow: { type: 'math_number', fields: { NUM: 0 } } }
                    }},
                    { kind: 'block', type: 'zigbee_wait', inputs: {
                        SECONDS: { shadow: { type: 'math_number', fields: { NUM: 5 } } }
                    }},
                    { kind: 'block', type: 'zigbee_log', inputs: {
                        MSG: { shadow: { type: 'text', fields: { TEXT: 'hello' } } }
                    }}
                ]
            },
            {
                kind: 'category',
                name: t('blockly.device_values'),
                colour: 210,
                contents: [
                    { kind: 'block', type: 'zigbee_event_value' },
                    { kind: 'block', type: 'zigbee_boolean' },
                    { kind: 'block', type: 'zigbee_get_property' }
                ]
            },
            {
                kind: 'category',
                name: t('blockly.datetime'),
                colour: 65,
                contents: [
                    { kind: 'block', type: 'system_datetime' },
                    { kind: 'block', type: 'system_time_between' }
                ]
            },
            {
                kind: 'category',
                name: t('blockly.notifications'),
                colour: 330,
                contents: [
                    { kind: 'block', type: 'telegram_send', inputs: {
                        MSG: { shadow: { type: 'text', fields: { TEXT: 'Alert!' } } }
                    }},
                    { kind: 'block', type: 'system_log', inputs: {
                        MSG: { shadow: { type: 'text', fields: { TEXT: 'something happened' } } }
                    }}
                ]
            },
            {
                kind: 'category',
                name: t('blockly.system'),
                colour: 0,
                contents: [
                    { kind: 'block', type: 'system_exec', inputs: {
                        CMD: { shadow: { type: 'text', fields: { TEXT: '/usr/bin/curl' } } }
                    }},
                    { kind: 'block', type: 'system_exec_no_output', inputs: {
                        CMD: { shadow: { type: 'text', fields: { TEXT: '/usr/bin/curl' } } }
                    }}
                ]
            },
            { kind: 'sep' },
            {
                kind: 'category',
                name: t('blockly.logic'),
                colour: 210,
                contents: [
                    { kind: 'block', type: 'controls_if' },
                    { kind: 'block', type: 'controls_if', extraState: { hasElse: true } },
                    { kind: 'block', type: 'logic_compare' },
                    { kind: 'block', type: 'logic_operation' },
                    { kind: 'block', type: 'logic_negate' },
                    { kind: 'block', type: 'logic_boolean' }
                ]
            },
            {
                kind: 'category',
                name: t('blockly.loops'),
                colour: 120,
                contents: [
                    { kind: 'block', type: 'controls_repeat_ext', inputs: {
                        TIMES: { shadow: { type: 'math_number', fields: { NUM: 10 } } }
                    }},
                    { kind: 'block', type: 'controls_whileUntil' },
                    { kind: 'block', type: 'controls_for' },
                    { kind: 'block', type: 'controls_forEach' }
                ]
            },
            {
                kind: 'category',
                name: t('blockly.math'),
                colour: 230,
                contents: [
                    { kind: 'block', type: 'math_number' },
                    { kind: 'block', type: 'math_arithmetic' },
                    { kind: 'block', type: 'math_modulo' },
                    { kind: 'block', type: 'math_constrain' },
                    { kind: 'block', type: 'math_random_int' }
                ]
            },
            {
                kind: 'category',
                name: t('blockly.text'),
                colour: 160,
                contents: [
                    { kind: 'block', type: 'text' },
                    { kind: 'block', type: 'text_join' },
                    { kind: 'block', type: 'text_length' },
                    { kind: 'block', type: 'text_indexOf' },
                    { kind: 'block', type: 'text_charAt' }
                ]
            },
            {
                kind: 'category',
                name: t('blockly.variables'),
                custom: 'VARIABLE',
                colour: 330
            }
        ]
    };
}

// === Dark Theme ===
var zigbeeDarkTheme = null;

function createDarkTheme() {
    if (zigbeeDarkTheme) return zigbeeDarkTheme;
    zigbeeDarkTheme = Blockly.Theme.defineTheme('zigbeeDark', {
        base: Blockly.Themes.Classic,
        componentStyles: {
            workspaceBackgroundColour: '#141416',
            toolboxBackgroundColour: '#1C1C1E',
            toolboxForegroundColour: '#E5E5E7',
            flyoutBackgroundColour: '#2C2C2E',
            flyoutForegroundColour: '#E5E5E7',
            flyoutOpacity: 0.95,
            scrollbarColour: '#3A3A3C',
            scrollbarOpacity: 0.5,
            insertionMarkerColour: '#4C8BF5',
            insertionMarkerOpacity: 0.5
        },
        fontStyle: {
            family: 'Inter, -apple-system, BlinkMacSystemFont, sans-serif',
            size: 12
        }
    });
    return zigbeeDarkTheme;
}

// === Editor lifecycle ===
async function openBlocklyEditor(scriptId) {
    currentScriptId = scriptId || null;

    // Load devices for dropdowns
    await loadDevices();

    var modal = document.getElementById('blockly-modal');
    modal.style.display = 'flex';

    // Define blocks and generators once
    if (!_blocksAndGensDefined) {
        defineBlocks();
        defineGenerators();
        _blocksAndGensDefined = true;
    }

    // Inject workspace if not already done
    if (!blocklyWorkspace) {
        blocklyWorkspace = Blockly.inject('blockly-div', {
            toolbox: buildToolbox(),
            theme: createDarkTheme(),
            grid: { spacing: 20, length: 3, colour: '#2C2C2E', snap: true },
            zoom: { controls: true, wheel: true, startScale: 1.0, maxScale: 3, minScale: 0.3 },
            trashcan: true,
            move: { scrollbars: true, drag: true, wheel: true }
        });

        // Live Lua preview update (debounced)
        blocklyWorkspace.addChangeListener(function() {
            debouncedLuaPreview();
        });

        // Resize workspace when window resizes
        window.addEventListener('resize', function() {
            if (blocklyWorkspace) {
                Blockly.svgResize(blocklyWorkspace);
            }
        });
    }

    // Load existing script data if editing
    if (scriptId) {
        try {
            var script = await apiCall("GET", "/api/automations/" + scriptId);
            document.getElementById('blockly-script-name').value = script.meta.name || '';
            document.getElementById('blockly-script-desc').value = script.meta.description || '';

            blocklyWorkspace.clear();
            if (script.blockly_xml) {
                var xml = Blockly.utils.xml.textToDom('<xml>' + script.blockly_xml + '</xml>');
                Blockly.Xml.domToWorkspace(xml, blocklyWorkspace);
            }
        } catch(e) {
            showToast(t("auto.toast.load_failed", e.message), true);
        }
    } else {
        document.getElementById('blockly-script-name').value = '';
        document.getElementById('blockly-script-desc').value = '';
        blocklyWorkspace.clear();
    }

    // Ensure layout is complete before resizing
    requestAnimationFrame(function() {
        Blockly.svgResize(blocklyWorkspace);
    });
}

function closeBlocklyEditor() {
    var modal = document.getElementById('blockly-modal');
    modal.style.display = 'none';
    currentScriptId = null;
}

async function saveBlocklyScript() {
    var name = document.getElementById('blockly-script-name').value.trim();
    if (!name) {
        showToast(t("auto.toast.enter_name"), true);
        return;
    }

    var xml = Blockly.Xml.workspaceToDom(blocklyWorkspace);
    var blocklyXml = Blockly.Xml.domToText(xml);
    // Strip outer <xml> wrapper for storage
    blocklyXml = blocklyXml.replace(/^<xml[^>]*>/, '').replace(/<\/xml>$/, '');

    var luaCode = Blockly.Lua.workspaceToCode(blocklyWorkspace);

    var data = {
        name: name,
        description: document.getElementById('blockly-script-desc').value.trim(),
        lua_code: luaCode,
        blockly_xml: blocklyXml,
        enabled: true
    };

    try {
        if (currentScriptId) {
            await apiCall("PUT", "/api/automations/" + currentScriptId, data);
            showToast(t("auto.toast.updated"));
        } else {
            await apiCall("POST", "/api/automations", data);
            showToast(t("auto.toast.created"));
        }
        closeBlocklyEditor();
        location.reload();
    } catch(e) {
        showToast(t("auto.toast.save_failed", e.message), true);
    }
}

function updateLuaPreview() {
    var pre = document.getElementById('lua-preview-code');
    if (!pre || !blocklyWorkspace) return;
    try {
        pre.textContent = Blockly.Lua.workspaceToCode(blocklyWorkspace);
    } catch(e) {
        pre.textContent = '-- Error generating Lua: ' + e.message;
    }
}

function debouncedLuaPreview() {
    if (_luaPreviewTimer) clearTimeout(_luaPreviewTimer);
    _luaPreviewTimer = setTimeout(updateLuaPreview, 300);
}

function toggleLuaPreview() {
    var panel = document.getElementById('lua-preview');
    if (!panel) return;
    var visible = panel.style.display !== 'none';
    panel.style.display = visible ? 'none' : 'flex';
    updateLuaPreview();
    if (blocklyWorkspace) Blockly.svgResize(blocklyWorkspace);
}

// === Page actions ===
function createAutomation() {
    openBlocklyEditor(null);
}

function editAutomation(id) {
    openBlocklyEditor(id);
}

async function toggleAutomation(id) {
    try {
        await apiCall("POST", "/api/automations/" + id + "/toggle");
        location.reload();
    } catch(e) {
        showToast(t("auto.toast.toggle_failed", e.message), true);
    }
}

async function runAutomation(id) {
    try {
        var result = await apiCall("POST", "/api/automations/" + id + "/run");
        if (result.ok) {
            var msg = t("auto.toast.run_ok", result.duration);
            if (result.logs && result.logs.length > 0) {
                msg += "\n" + result.logs.join("\n");
            }
            showToast(msg);
        } else {
            showToast(t("auto.toast.run_failed", result.error), true);
        }
    } catch(e) {
        showToast(t("auto.toast.run_request_failed", e.message), true);
    }
}

async function runBlocklyScript() {
    if (!blocklyWorkspace) return;
    var luaCode;
    try {
        luaCode = Blockly.Lua.workspaceToCode(blocklyWorkspace);
    } catch(e) {
        showToast(t("auto.toast.run_failed", e.message), true);
        return;
    }
    if (!luaCode || !luaCode.trim()) {
        showToast(t("auto.toast.run_failed", "empty script"), true);
        return;
    }
    try {
        var result = await apiCall("POST", "/api/automations/_inline/run", { lua_code: luaCode });
        if (result.ok) {
            var msg = t("auto.toast.run_ok", result.duration);
            if (result.logs && result.logs.length > 0) {
                msg += "\n" + result.logs.join("\n");
            }
            showToast(msg);
        } else {
            showToast(t("auto.toast.run_failed", result.error), true);
        }
    } catch(e) {
        showToast(t("auto.toast.run_request_failed", e.message), true);
    }
}

async function deleteAutomation(id) {
    if (!confirm(t("auto.toast.delete_confirm"))) return;
    try {
        await apiCall("DELETE", "/api/automations/" + id);
        location.reload();
    } catch(e) {
        showToast(t("auto.toast.delete_failed", e.message), true);
    }
}
