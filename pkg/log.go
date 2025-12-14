package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	apiURL            string
	logTelegramChatID string
)

// TelegramHook implements logrus.Hook for sending logs to Telegram
type TelegramHook struct {
	client *http.Client
	level  logrus.Level
}

// AppNameHook injects a static app_name field into all logs
type AppNameHook struct {
	AppName string
}

func (h *AppNameHook) Levels() []logrus.Level { return logrus.AllLevels }

func (h *AppNameHook) Fire(entry *logrus.Entry) error {
	if _, ok := entry.Data["app_name"]; !ok {
		entry.Data["app_name"] = h.AppName
	}
	return nil
}

// --- Init global logger ---
func init() {
	logrus.SetReportCaller(false)

	level, err := logrus.ParseLevel(EnvOrDefault("LOG_LEVEL", "info"))
	if err != nil {
		logrus.WithError(err).Panic("invalid LOG_LEVEL")
	}
	logrus.SetLevel(level)

	appName := EnvOrDefault("APP_NAME", "__MODULE__")
	logrus.AddHook(&AppNameHook{AppName: appName})
	logrus.AddHook(NewMaskingHook())

	if enabled, _ := strconv.ParseBool(EnvOrDefault("TELEGRAM_LOG_ENABLE", "false")); enabled {
		hook, err := newTelegramHook(logrus.WarnLevel)
		if err != nil {
			logrus.WithError(err).Error("failed to create Telegram hook")
		} else {
			logrus.AddHook(hook)
		}
	}

	logrus.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: true,
		TimestampFormat: func() string {
			return "2006-01-02 15:04:05"
		}(),
	})
}

// --- Panic-safe wrapper ---
func SafeRun(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithField("panic", r).Error("application panicked")
			// Optionally re-panic if you want crash instead of silent recover:
			panic(r)
		}
	}()
	fn()
}

// --- Telegram Hook ---

func newTelegramHook(level logrus.Level) (*TelegramHook, error) {
	token := EnvOrDefault("LOG_TELEGRAM_BOT_TOKEN", "")
	if token == "" {
		return nil, errors.New("LOG_TELEGRAM_BOT_TOKEN is empty")
	}

	logTelegramChatID = EnvOrDefault("LOG_TELEGRAM_CHAT_ID", "")
	if logTelegramChatID == "" {
		return nil, errors.New("LOG_TELEGRAM_CHAT_ID is empty")
	}

	apiURL = "https://api.telegram.org/bot" + token + "/sendMessage"
	client := &http.Client{Timeout: 3 * time.Second}

	return &TelegramHook{client: client, level: level}, nil
}

func (h *TelegramHook) Levels() []logrus.Level { return logrus.AllLevels }

func (h *TelegramHook) Fire(entry *logrus.Entry) error {

	msgBytes, _ := entry.Bytes() // همون فرمتری که بالا set کردی (JSONFormatter) رو استفاده می‌کنه
	msg := string(msgBytes)

	msg = strings.ReplaceAll(msg, "`", "\\`")

	codeBlock := "```json\n" + msg + "\n```"

	payload := map[string]interface{}{
		"chat_id":    logTelegramChatID,
		"text":       codeBlock,
		"parse_mode": "MarkdownV2",
	}

	body, _ := json.Marshal(payload)

	if entry.Level == logrus.PanicLevel || entry.Level == logrus.FatalLevel {
		h.sendTelegramSync(body)
	} else {
		// for warn/error we can fire-and-forget
		go h.sendTelegram(body)
	}
	return nil
}

// (within the same package)
func (h *TelegramHook) sendTelegram(body []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		jsonInternalLog("warning", "TelegramHook: failed to create request")
		return
	}

	// Make body rewindable so http2 transport can retry without "Request.Body was written" error
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		jsonInternalLog("warning", "TelegramHook: send failed")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if b, _ := io.ReadAll(resp.Body); len(b) > 0 {
			jsonInternalLog("warning", "TelegramHook: Telegram API error")
		} else {
			jsonInternalLog("warning", "TelegramHook: Telegram API error")
		}
	}
}

// --- Zero-allocation reader ---
type bytesReaderPool struct {
	buf []byte
}

func bytesReader(b []byte) io.Reader {
	// Use simple wrapper instead of bytes.NewBuffer (saves ~3 allocs)
	return &bytesReaderPool{buf: b}
}

func (r *bytesReaderPool) Read(p []byte) (int, error) {
	if len(r.buf) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.buf)
	r.buf = r.buf[n:]
	return n, nil
}

// synchronous send used for Panic/Fatal
func (h *TelegramHook) sendTelegramSync(body []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytesReader(body))
	if err != nil {
		// use stdlib to avoid recursion into logrus hooks
		fmt.Fprintf(os.Stderr, "TelegramHook(sync): failed to create request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "TelegramHook(sync): send failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if b, _ := io.ReadAll(resp.Body); len(b) > 0 {
			fmt.Fprintf(os.Stderr, "TelegramHook(sync): Telegram API error (%d): %s\n", resp.StatusCode, string(b))
		}
	}
}

// MaskingHook masks sensitive values in entry.Message and entry.Data
type MaskingHook struct {
	patterns []*regexp.Regexp
	repls    []string
}

// NewMaskingHook returns a configured masking hook.
// You can easily extend regexes/repls here.
func NewMaskingHook() *MaskingHook {
	// Order matters — token URL and bearer first, then key=value patterns, then generic patterns.
	patterns := []*regexp.Regexp{
		// telegram bot token in URL: api.telegram.org/bot<token>
		regexp.MustCompile(`(api\.telegram\.org\/bot)([A-Za-z0-9:_-]{8,})`),

		// Authorization: Bearer <token>
		regexp.MustCompile(`(?i)(authorization\s*:\s*bearer\s+)(\S+)`),

		// key=value sequences (postgres style): host=..., user=..., password=..., database=..., dbname=...
		// matches e.g. host=localhost user=postgres password=secret
		regexp.MustCompile(`(?i)\b(host|user|password|database|dbname)\s*=\s*[^\s,}]+`),

		// typical key: value or key="value" or key='value' patterns (api_key, token, access_token, secret, client_secret, authorization)
		regexp.MustCompile(`(?i)\b(api_key|apikey|token|access_token|secret|client_secret|authorization|auth)\s*[:=]\s*(?:".*?"|'.*?'|[^\s,}]+)`),

		// long hex / base-like tokens (32+ hex chars)
		regexp.MustCompile(`(?i)[A-Fa-f0-9]{32,}`),
	}

	repls := []string{
		`${1}<redacted>`, // telegram URL -> keep prefix, redact token
		`$1<redacted>`,   // authorization: bearer -> keep prefix, redact
		`$0`,             // placeholder: we'll handle this specially (see below)
		`$0`,             // placeholder for generic key groups (handle by function)
		`<redacted>`,     // long hex -> redact entirely
	}

	// For the placeholder patterns (3 and 4) we will not use the simple ReplaceAllString with $0,
	// instead we'll perform custom replacements in maskString to preserve the key and replace only the value.
	return &MaskingHook{patterns: patterns, repls: repls}
}

func (h *MaskingHook) Levels() []logrus.Level { return logrus.AllLevels }

func (h *MaskingHook) Fire(entry *logrus.Entry) error {
	// Mask the entry message
	entry.Message = h.maskString(entry.Message)

	// Mask all data values — convert everything to string and replace with masked strings.
	for k, v := range entry.Data {
		// Convert to string (keeps JSON formatting predictable)
		str := fmt.Sprint(v)
		entry.Data[k] = h.maskString(str)
	}
	return nil
}

// maskString applies the regex-based masking rules
func (h *MaskingHook) maskString(s string) string {
	if s == "" {
		return s
	}

	// 1) direct replacements for first two simple patterns (they have direct $1 or \1 usage)
	s = h.patterns[0].ReplaceAllString(s, h.repls[0]) // telegram url token
	s = h.patterns[1].ReplaceAllString(s, h.repls[1]) // bearer token

	// 2) postgres-style key=value: replace `key=value` -> `key=<redacted>`
	// using a regex that captures key and replaces the whole match with "key=<redacted>"
	kvRe := regexp.MustCompile(`(?i)\b(host|user|password|database|dbname)\s*=\s*([^\s,}]+)`)
	s = kvRe.ReplaceAllString(s, "${1}=<redacted>")

	// 3) generic key: value or key="value" or key='value'
	// capture the key and separator and replace the value with <redacted>
	genRe := regexp.MustCompile(`(?i)\b(api_key|apikey|token|access_token|secret|client_secret|authorization|auth)\s*([:=])\s*(".*?"|'.*?'|[^\s,}]+)`)
	s = genRe.ReplaceAllString(s, "${1}${2}<redacted>")

	// 4) redact long hex-like tokens (32+ hex chars)
	s = h.patterns[4].ReplaceAllString(s, h.repls[4])

	return s
}

// change to your app name constant
const appName = "__MODULE__"

// jsonInternalLog writes a single-line JSON log to stderr (won't trigger logrus hooks)
func jsonInternalLog(level, msg string) {
	// caller info
	pc, file, line, ok := runtime.Caller(2) // 2 to point to caller of helper
	funcName := ""
	if ok {
		if f := runtime.FuncForPC(pc); f != nil {
			funcName = f.Name()
		}
	}

	// base payload matching your format
	payload := map[string]interface{}{
		"app_name": appName,
		"file":     file + ":" + itoa(line),
		"func":     funcName,
		"level":    level,
		"msg":      msg,
		"time":     time.Now().Format("2006-01-02 15:04:05"),
	}

	enc, _ := json.Marshal(payload) // ignore error — best-effort internal logging
	os.Stderr.Write(append(enc, '\n'))
}

// tiny itoa for line numbers (avoids importing strconv repeatedly)
func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
