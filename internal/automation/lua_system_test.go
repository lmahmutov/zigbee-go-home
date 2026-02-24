//go:build !no_automation

package automation

import (
	"log/slog"
	"os"
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func newTestEngine() *Engine {
	return &Engine{
		systemCfg:   SystemConfig{},
		telegramCfg: TelegramConfig{},
	}
}

func TestSystemDatetimeReturnsNumber(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	e := newTestEngine()
	registerSystemModule(L, e)

	numberComponents := []string{"hour", "minute", "second", "weekday", "day", "month", "year", "timestamp"}
	for _, comp := range numberComponents {
		L.SetGlobal("_comp", lua.LString(comp))
		if err := L.DoString(`_result = system.datetime(_comp)`); err != nil {
			t.Fatalf("system.datetime(%q) error: %v", comp, err)
		}
		result := L.GetGlobal("_result")
		if result.Type() != lua.LTNumber {
			t.Errorf("system.datetime(%q) type = %v, want LTNumber", comp, result.Type())
		}
	}
}

func TestSystemDatetimeReturnsString(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	e := newTestEngine()
	registerSystemModule(L, e)

	stringComponents := []string{"time_str", "date_str"}
	for _, comp := range stringComponents {
		L.SetGlobal("_comp", lua.LString(comp))
		if err := L.DoString(`_result = system.datetime(_comp)`); err != nil {
			t.Fatalf("system.datetime(%q) error: %v", comp, err)
		}
		result := L.GetGlobal("_result")
		if result.Type() != lua.LTString {
			t.Errorf("system.datetime(%q) type = %v, want LTString", comp, result.Type())
		}
	}
}

func TestSystemDatetimeHourRange(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	e := newTestEngine()
	registerSystemModule(L, e)

	if err := L.DoString(`_hour = system.datetime("hour")`); err != nil {
		t.Fatal(err)
	}
	hour := int(L.GetGlobal("_hour").(lua.LNumber))
	if hour < 0 || hour > 23 {
		t.Errorf("hour = %d, want 0-23", hour)
	}
}

func TestSystemTimeBetweenNormalRange(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	e := newTestEngine()
	registerSystemModule(L, e)

	hour := time.Now().Hour()

	// Test a range that includes the current hour
	from := hour
	to := hour + 1
	if to > 23 {
		to = 0
	}

	L.SetGlobal("_from", lua.LNumber(from))
	L.SetGlobal("_to", lua.LNumber(to))
	if err := L.DoString(`_result = system.time_between(_from, _to)`); err != nil {
		t.Fatal(err)
	}
	result := L.GetGlobal("_result")
	if result != lua.LTrue {
		t.Errorf("time_between(%d, %d) at hour %d = false, want true", from, to, hour)
	}
}

func TestSystemTimeBetweenMidnightWrap(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	e := newTestEngine()
	registerSystemModule(L, e)

	hour := time.Now().Hour()

	// Create a midnight-wrapping range that includes the current hour
	// If hour is 14, use range 10-4 (which wraps: 10,11,...23,0,1,2,3)
	from := hour - 4
	if from < 0 {
		from += 24
	}
	to := hour - 8
	if to < 0 {
		to += 24
	}
	// This makes from > to (midnight wrap), and current hour is within range

	L.SetGlobal("_from", lua.LNumber(from))
	L.SetGlobal("_to", lua.LNumber(to))
	if err := L.DoString(`_result = system.time_between(_from, _to)`); err != nil {
		t.Fatal(err)
	}
	result := L.GetGlobal("_result")
	if result != lua.LTrue {
		t.Errorf("time_between(%d, %d) at hour %d = false, want true (midnight wrap)", from, to, hour)
	}
}

func TestSystemTimeBetweenOutsideRange(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	e := newTestEngine()
	registerSystemModule(L, e)

	hour := time.Now().Hour()

	// Create a normal range that does NOT include the current hour
	from := hour + 2
	if from > 23 {
		from -= 24
	}
	to := hour + 4
	if to > 23 {
		to -= 24
	}
	// Ensure normal range (from < to)
	if from > to {
		from, to = to, from
	}
	// Make sure current hour is outside [from, to)
	if hour >= from && hour < to {
		// Shift both
		from = hour + 3
		if from > 23 {
			from -= 24
		}
		to = hour + 5
		if to > 23 {
			to -= 24
		}
		if from > to {
			from, to = to, from
		}
	}

	L.SetGlobal("_from", lua.LNumber(from))
	L.SetGlobal("_to", lua.LNumber(to))
	if err := L.DoString(`_result = system.time_between(_from, _to)`); err != nil {
		t.Fatal(err)
	}
	result := L.GetGlobal("_result")
	if result != lua.LFalse {
		t.Errorf("time_between(%d, %d) at hour %d = true, want false", from, to, hour)
	}
}

func TestSystemExecBlockedWhenAllowlistEmpty(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	e := newTestEngine()
	e.logger = testLogger()
	registerSystemModule(L, e)

	if err := L.DoString(`_result = system.exec("ls")`); err != nil {
		t.Fatal(err)
	}
	result := L.GetGlobal("_result")
	if s, ok := result.(lua.LString); !ok || string(s) != "" {
		t.Errorf("exec with empty allowlist returned %q, want empty string", result)
	}
}

func TestSystemExecBlockedNotInAllowlist(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	e := newTestEngine()
	e.logger = testLogger()
	e.systemCfg.ExecAllowlist = []string{"/usr/bin/echo"}
	registerSystemModule(L, e)

	if err := L.DoString(`_result = system.exec("/usr/bin/ls")`); err != nil {
		t.Fatal(err)
	}
	result := L.GetGlobal("_result")
	if s, ok := result.(lua.LString); !ok || string(s) != "" {
		t.Errorf("exec with non-allowlisted cmd returned %q, want empty string", result)
	}
}

func TestSystemExecAllowed(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	e := newTestEngine()
	e.logger = testLogger()
	e.systemCfg.ExecAllowlist = []string{"/bin/echo"}
	e.systemCfg.ExecTimeout = 5 * time.Second
	registerSystemModule(L, e)

	if err := L.DoString(`_result = system.exec("/bin/echo hello")`); err != nil {
		t.Fatal(err)
	}
	result := L.GetGlobal("_result")
	s, ok := result.(lua.LString)
	if !ok {
		t.Fatalf("exec returned type %v, want LTString", result.Type())
	}
	if string(s) != "hello\n" {
		t.Errorf("exec returned %q, want %q", string(s), "hello\n")
	}
}

func TestTelegramSendNoConfig(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	e := newTestEngine()
	e.logger = testLogger()
	registerTelegramModule(L, e)

	// Should not panic with empty config
	if err := L.DoString(`telegram.send("test")`); err != nil {
		t.Fatal(err)
	}
}
