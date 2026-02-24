//go:build !no_automation

package automation

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// SystemConfig holds configuration for the system Lua module.
type SystemConfig struct {
	ExecAllowlist []string      // allowed command paths
	ExecTimeout   time.Duration // timeout for exec commands
}

// TelegramConfig holds configuration for the telegram Lua module.
type TelegramConfig struct {
	BotToken string
	ChatIDs  []string
}

// registerSystemModule registers the `system` global table in a Lua state.
func registerSystemModule(L *lua.LState, e *Engine) {
	mod := L.NewTable()

	mod.RawSetString("datetime", L.NewFunction(func(L *lua.LState) int {
		return systemDatetime(L)
	}))

	mod.RawSetString("time_between", L.NewFunction(func(L *lua.LState) int {
		return systemTimeBetween(L)
	}))

	mod.RawSetString("log", L.NewFunction(func(L *lua.LState) int {
		return systemLog(L, e)
	}))

	mod.RawSetString("exec", L.NewFunction(func(L *lua.LState) int {
		return systemExec(L, e)
	}))

	L.SetGlobal("system", mod)
}

// registerTelegramModule registers the `telegram` global table in a Lua state.
func registerTelegramModule(L *lua.LState, e *Engine) {
	mod := L.NewTable()

	mod.RawSetString("send", L.NewFunction(func(L *lua.LState) int {
		return telegramSend(L, e)
	}))

	L.SetGlobal("telegram", mod)
}

// system.datetime(component) — returns a date/time component
func systemDatetime(L *lua.LState) int {
	component := L.CheckString(1)
	now := time.Now()

	switch component {
	case "hour":
		L.Push(lua.LNumber(now.Hour()))
	case "minute":
		L.Push(lua.LNumber(now.Minute()))
	case "second":
		L.Push(lua.LNumber(now.Second()))
	case "weekday":
		L.Push(lua.LNumber(now.Weekday()))
	case "day":
		L.Push(lua.LNumber(now.Day()))
	case "month":
		L.Push(lua.LNumber(now.Month()))
	case "year":
		L.Push(lua.LNumber(now.Year()))
	case "timestamp":
		L.Push(lua.LNumber(now.Unix()))
	case "time_str":
		L.Push(lua.LString(now.Format("15:04:05")))
	case "date_str":
		L.Push(lua.LString(now.Format("2006-01-02")))
	default:
		L.ArgError(1, "unknown component: "+component)
		return 0
	}
	return 1
}

// system.time_between(from_hour, to_hour) — checks if current hour is in range (supports midnight wrap)
func systemTimeBetween(L *lua.LState) int {
	from := L.CheckInt(1)
	to := L.CheckInt(2)
	hour := time.Now().Hour()

	var result bool
	if from <= to {
		// Normal range: e.g. 8-22
		result = hour >= from && hour < to
	} else {
		// Midnight-wrapping range: e.g. 22-6
		result = hour >= from || hour < to
	}

	L.Push(lua.LBool(result))
	return 1
}

// system.log(level, msg)
func systemLog(L *lua.LState, e *Engine) int {
	level := L.CheckString(1)
	msg := L.CheckString(2)

	switch level {
	case "debug":
		e.logger.Debug("script log", "msg", msg)
	case "warn":
		e.logger.Warn("script log", "msg", msg)
	case "error":
		e.logger.Error("script log", "msg", msg)
	default:
		e.logger.Info("script log", "msg", msg)
	}
	return 0
}

// system.exec(cmd) — execute an allowlisted command, return stdout
func systemExec(L *lua.LState, e *Engine) int {
	cmdStr := L.CheckString(1)

	// Parse command into binary + args
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		L.ArgError(1, "empty command")
		return 0
	}
	binary := parts[0]

	// Require absolute path
	if !filepath.IsAbs(binary) {
		e.logger.Warn("exec blocked: not an absolute path", "cmd", binary)
		L.Push(lua.LString(""))
		return 1
	}

	// Check allowlist
	allowed := false
	for _, a := range e.systemCfg.ExecAllowlist {
		if a == binary {
			allowed = true
			break
		}
	}
	if !allowed {
		e.logger.Warn("exec blocked: not in allowlist", "cmd", binary)
		L.Push(lua.LString(""))
		return 1
	}

	timeout := e.systemCfg.ExecTimeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, parts[1:]...)
	// Cap stdout at 64KB + 1 to detect overflow
	stdout, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			e.logger.Warn("exec timeout", "cmd", binary, "timeout", timeout)
		} else {
			e.logger.Warn("exec failed", "cmd", binary, "err", err)
		}
		L.Push(lua.LString(""))
		return 1
	}

	// Cap output at 64KB
	if len(stdout) > 65536 {
		stdout = stdout[:65536]
	}

	L.Push(lua.LString(string(stdout)))
	return 1
}

// telegram.send(msg) — send message to all configured chat IDs
func telegramSend(L *lua.LState, e *Engine) int {
	msg := L.CheckString(1)

	if e.telegramCfg.BotToken == "" {
		e.logger.Warn("telegram.send: bot_token not configured")
		return 0
	}

	if len(e.telegramCfg.ChatIDs) == 0 {
		e.logger.Warn("telegram.send: no chat_ids configured")
		return 0
	}

	// Fire-and-forget for each chat ID
	for _, chatID := range e.telegramCfg.ChatIDs {
		go func(cid string) {
			url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", e.telegramCfg.BotToken)
			body := fmt.Sprintf(`{"chat_id":%q,"text":%q}`, cid, msg)

			req, err := http.NewRequest("POST", url, strings.NewReader(body))
			if err != nil {
				e.logger.Error("telegram request create", "err", err)
				return
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				e.logger.Error("telegram send", "err", err, "chat_id", cid)
				return
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				e.logger.Warn("telegram send non-200", "status", resp.StatusCode, "chat_id", cid)
			}
		}(chatID)
	}

	return 0
}
