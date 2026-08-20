package platform

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var ErrNotFound = errors.New("resource not found")
var ErrConflict = errors.New("resource conflict")
var ErrInvalid = errors.New("invalid request")
var ErrQuota = errors.New("quota exceeded")

type Clock interface{ Now() time.Time }
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now().UTC() }

type Config struct {
	HTTPAddr    string `json:"http_addr"`
	GRPCAddr    string `json:"grpc_addr"`
	DataDir     string `json:"data_dir"`
	QueryBudget int    `json:"query_budget"`
	MaxBody     int64  `json:"max_body"`
	LogLevel    string `json:"log_level"`
}

func DefaultConfig() Config {
	return Config{HTTPAddr: ":8086", GRPCAddr: ":9096", DataDir: "./data", QueryBudget: 10000, MaxBody: 4 << 20, LogLevel: "info"}
}
func LoadConfig(path string) (Config, error) {
	c := DefaultConfig()
	b, e := os.ReadFile(path)
	if e == nil {
		if e = json.Unmarshal(b, &c); e != nil {
			if e = parseYAMLConfig(string(b), &c); e != nil {
				return c, fmt.Errorf("decode config: %w", e)
			}
		}
	}
	if v := os.Getenv("SEARCH_HTTP_ADDR"); v != "" {
		c.HTTPAddr = v
	}
	if v := os.Getenv("SEARCH_DATA_DIR"); v != "" {
		c.DataDir = v
	}
	if v := os.Getenv("SEARCH_QUERY_BUDGET"); v != "" {
		fmt.Sscanf(v, "%d", &c.QueryBudget)
	}
	if v := os.Getenv("SEARCH_GRPC_ADDR"); v != "" {
		c.GRPCAddr = v
	}
	if v := os.Getenv("SEARCH_MAX_BODY"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			c.MaxBody = n
		}
	}
	if v := os.Getenv("SEARCH_LOG_LEVEL"); v != "" {
		c.LogLevel = v
	}
	return c, nil
}

func parseYAMLConfig(text string, c *Config) error {
	s := bufio.NewScanner(strings.NewReader(text))
	for s.Scan() {
		parts := strings.SplitN(strings.TrimSpace(s.Text()), ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
		switch key {
		case "http_addr":
			c.HTTPAddr = value
		case "grpc_addr":
			c.GRPCAddr = value
		case "data_dir":
			c.DataDir = value
		case "query_budget":
			n, err := strconv.Atoi(value)
			if err != nil {
				return err
			} else {
				c.QueryBudget = n
			}
		case "max_body":
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			} else {
				c.MaxBody = n
			}
		case "log_level":
			c.LogLevel = value
		}
	}
	return s.Err()
}

type Logger struct{ l *slog.Logger }

func NewLogger(level string) *Logger {
	var lv slog.Level
	if level == "debug" {
		lv = slog.LevelDebug
	}
	return &Logger{slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lv}))}
}
func (l *Logger) Info(msg string, args ...any)  { l.l.Info(msg, args...) }
func (l *Logger) Error(msg string, args ...any) { l.l.Error(msg, args...) }
func (l *Logger) With(args ...any) *Logger      { return &Logger{l.l.With(args...)} }

type EventLog struct {
	mu   sync.Mutex
	path string
}

func (w *EventLog) AppendBatchWithWriter(ctx context.Context, events []any, open func() (io.WriteCloser, error)) error {
	for _, event := range events {
		f, e := open()
		if e != nil {
			return e
		}
		defer f.Close()
		b, e := json.Marshal(event)
		if e != nil {
			return fmt.Errorf("marshal event: %w", e)
		}
		if _, e = f.Write(append(b, '\n')); e != nil {
			return fmt.Errorf("write batch event: %w", e)
		}
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func NewEventLog(dir string) (*EventLog, error) {
	if e := os.MkdirAll(dir, 0755); e != nil {
		return nil, e
	}
	return &EventLog{path: filepath.Join(dir, "wal.jsonl")}, nil
}
func (w *EventLog) Append(ctx context.Context, event any) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	b, e := json.Marshal(event)
	if e != nil {
		return fmt.Errorf("marshal event: %w", e)
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	f, e := os.OpenFile(w.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if e != nil {
		return fmt.Errorf("open wal: %w", e)
	}
	defer f.Close()
	if _, e = f.Write(append(b, '\n')); e != nil {
		return fmt.Errorf("write wal: %w", e)
	}
	return nil
}
func (w *EventLog) ReadAll() ([]json.RawMessage, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	b, e := os.ReadFile(w.path)
	if os.IsNotExist(e) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	var out []json.RawMessage
	for _, line := range splitLines(b) {
		if len(line) > 0 {
			out = append(out, json.RawMessage(line))
		}
	}
	return out, nil
}
func splitLines(b []byte) [][]byte {
	var r [][]byte
	start := 0
	for i, c := range b {
		if c == '\n' {
			r = append(r, b[start:i])
			start = i + 1
		}
	}
	if start < len(b) {
		r = append(r, b[start:])
	}
	return r
}

type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Quota     int       `json:"quota"`
	CreatedAt time.Time `json:"created_at"`
}
type FieldType string

const (
	Text      FieldType = "text"
	Keyword   FieldType = "keyword"
	Integer   FieldType = "integer"
	Float     FieldType = "float"
	Boolean   FieldType = "boolean"
	Timestamp FieldType = "timestamp"
	GeoPoint  FieldType = "geo_point"
)

type FieldMapping struct {
	Name     string    `json:"name"`
	Type     FieldType `json:"type"`
	Analyzer string    `json:"analyzer,omitempty"`
	Required bool      `json:"required,omitempty"`
}
type Collection struct {
	ID             string                  `json:"id"`
	TenantID       string                  `json:"tenant_id"`
	Name           string                  `json:"name"`
	Mappings       map[string]FieldMapping `json:"mappings"`
	MappingVersion int                     `json:"mapping_version"`
	Shards         int                     `json:"shards"`
	Replicas       int                     `json:"replicas"`
	ReadOnly       bool                    `json:"read_only"`
	CreatedAt      time.Time               `json:"created_at"`
}
type Document struct {
	ID           string         `json:"id"`
	TenantID     string         `json:"tenant_id"`
	CollectionID string         `json:"collection_id"`
	Data         map[string]any `json:"data"`
	Version      int64          `json:"version"`
	ExternalTS   int64          `json:"external_ts,omitempty"`
	Deleted      bool           `json:"deleted"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
type Token struct {
	Term  string
	Pos   int
	Start int
	End   int
}
type Posting struct {
	DocID     string `json:"doc_id"`
	Frequency int    `json:"frequency"`
	Positions []int  `json:"positions"`
}
type SearchHit struct {
	ID        string              `json:"id"`
	Score     float64             `json:"score"`
	Source    map[string]any      `json:"source"`
	Highlight map[string][]string `json:"highlight,omitempty"`
}
type SearchResult struct {
	Hits         []SearchHit               `json:"hits"`
	Total        int                       `json:"total"`
	NextCursor   string                    `json:"next_cursor,omitempty"`
	TookMS       int64                     `json:"took_ms"`
	Aggregations map[string]map[string]int `json:"aggregations,omitempty"`
}
