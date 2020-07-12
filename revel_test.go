package revel

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	log15 "github.com/inconshreveable/log15"
	"github.com/revel/revel"
	rlogger "github.com/revel/revel/logger"
	"github.com/stretchr/testify/assert"
	"gopkg.in/birkirb/loggers.v1"
	"gopkg.in/birkirb/loggers.v1/log"
)

func TestRevelInterface(t *testing.T) {
	var _ loggers.Contextual = NewLogger()
}

func TestRevelLevelOutputWithColor(t *testing.T) {
	l, b := newBufferedRevelLog()
	l.Debugln("\x1b[30mThis text will have black color\x1b[0m")
	l.Debugln("This text will have default color")
	var expectedMatch = []string{
		"TRACE.*This text will have black color.+$",
		"TRACE.*This text will have default color",
	}
	actual := b.String()
	lines := strings.Split(actual, "\n")
	k := 1 // Offset for lines before expected

	for i, expected := range expectedMatch {
		if ok, _ := regexp.Match(expected, []byte(lines[i+k])); !ok {
			t.Errorf("Log output mismatch: `%s` (actual) != `%s` (expected)", lines[i+k], expected)
		}
	}
}

func TestRevelLevelOutput(t *testing.T) {
	l, b := newBufferedRevelLog()
	l.Info("This is a test")

	expectedMatch := "INFO.*This is a test\n"
	actual := b.String()
	if ok, _ := regexp.Match(expectedMatch, []byte(actual)); !ok {
		t.Errorf("Log output mismatch: %s (actual) != %s (expected)", actual, expectedMatch)
	}
}

func TestRevelLevelfOutput(t *testing.T) {
	l, b := newBufferedRevelLog()
	l.Errorf("This is %s test", "a")

	expectedMatch := "ERROR.*This is a test\n"
	actual := b.String()
	if ok, _ := regexp.Match(expectedMatch, []byte(actual)); !ok {
		t.Errorf("Log output mismatch: %s (actual) != %s (expected)", actual, expectedMatch)
	}
}

func TestRevelLevellnOutput(t *testing.T) {
	l, b := newBufferedRevelLog()
	l.Debugln("This is a test.", "So is this.")

	expectedMatch := "DBUG.*This is a test. So is this.\n"
	actual := b.String()
	assert.Equal(t, expectedMatch, actual)
}

func TestRevelWithFieldsOutput(t *testing.T) {
	l, b := newBufferedRevelLog()
	l.WithFields("test", true).Warn("This is a message.")

	expectedMatch := "WARN.*This is a message. test=true\n"
	actual := b.String()
	assert.Equal(t, expectedMatch, actual)
}

func TestRevelWithFieldsfOutput(t *testing.T) {
	l, b := newBufferedRevelLog()
	l.WithFields("test", true, "Error", "serious").Errorf("This is a %s.", "message")

	expectedMatch := "EROR.*This is a message.   test=true Error=serious\n"
	actual := b.String()
	assert.Equal(t, expectedMatch, actual)
}

var lvlMap = map[int]string{
	0: "critical",
	1: "error",
	2: "warn",
	3: "info",
	4: "debug",
}

func logfmt(buf *bytes.Buffer, ctx []interface{}, color int) {
	for i := 0; i < len(ctx); i += 2 {
		if i != 0 {
			buf.WriteByte(' ')
		}

		k := ctx[i].(string)
		v := ctx[i+1].(string)

		// XXX: we should probably check that all of your key bytes aren't invalid
		if color > 0 {
			fmt.Fprintf(buf, "\x1b[%dm%s\x1b[0m=%s", color, k, v)
		} else {
			buf.WriteString(k)
			buf.WriteByte('=')
			buf.WriteString(v)
		}
	}

	buf.WriteByte('\n')
}
func FormatTest() log15.Format {
	return log15.FormatFunc(func(r *log15.Record) []byte {
		var color = 0
		switch r.Lvl {
		case log15.LvlCrit:
			color = 35
		case log15.LvlError:
			color = 31
		case log15.LvlWarn:
			color = 33
		case log15.LvlInfo:
			color = 32
		case log15.LvlDebug:
			color = 36
		}

		b := &bytes.Buffer{}
		lvl := strings.ToUpper(r.Lvl.String())
		if color > 0 {
			//fmt.Fprintf(b, "\x1b[%dm%s\x1b[0m %s ", color, lvl, r.Msg)
			fmt.Fprintf(b, "%s.*%s", lvl, r.Msg)
		} else {
			fmt.Fprintf(b, "[%s] %s", lvl, r.Msg)
		}

		// print the keys logfmt style
		logfmt(b, r.Ctx, color)
		return b.Bytes()
	})
}

type LogHandler struct {
	h log15.Handler
}

func newLogHandler(w io.Writer) *LogHandler {
	return &LogHandler{
		h: log15.StreamHandler(w, FormatTest()),
	}
}

func (h *LogHandler) Log(r *rlogger.Record) error {
	lvl, err := log15.LvlFromString(lvlMap[int(r.Level)])
	if err != nil {
		panic(err)
	}
	return h.h.Log(&log15.Record{
		Msg: r.Message,
		Lvl: lvl,
	})
}

func newBufferedRevelLog() (loggers.Contextual, *bytes.Buffer) {
	var b []byte
	var bb = bytes.NewBuffer(b)
	h := newLogHandler(bb)

	revel.RootLog = rlogger.New()
	revel.AppLog.SetHandler(rlogger.FuncHandler(h.Log))

	return NewLogger(), bb
}

func TestBackTrace(t *testing.T) {
	l, b := newBufferedRevelLog()
	log.Logger = l
	log.Error("an error")
	_, file, line, _ := runtime.Caller(0)

	mustContain := fmt.Sprintf("%s:%d", filepath.Base(file), line-1)
	actual := b.String()
	if ok := strings.Contains(actual, mustContain); !ok {
		t.Errorf("Log output mismatch: %s (actual) != %s (expected)", actual, mustContain)
	}
}

func TestBackTraceF(t *testing.T) {
	l, b := newBufferedRevelLog()
	log.Logger = l
	log.Errorf("an error: %s", "value")
	_, file, line, _ := runtime.Caller(0)

	mustContain := fmt.Sprintf("%s:%d", filepath.Base(file), line-1)
	actual := b.String()
	if ok := strings.Contains(actual, mustContain); !ok {
		t.Errorf("Log output mismatch: %s (actual) != %s (expected)", actual, mustContain)
	}
}

func TestBackTraceLn(t *testing.T) {
	l, b := newBufferedRevelLog()
	log.Logger = l
	log.Errorln("an error")
	_, file, line, _ := runtime.Caller(0)

	mustContain := fmt.Sprintf("%s:%d", filepath.Base(file), line-1)
	actual := b.String()
	if ok := strings.Contains(actual, mustContain); !ok {
		t.Errorf("Log output mismatch: %s (actual) != %s (expected)", actual, mustContain)
	}
}
