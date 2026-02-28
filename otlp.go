package trace_otlp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bamgoo/bamgoo"
	. "github.com/bamgoo/base"
	"github.com/bamgoo/trace"
)

type (
	otlpDriver struct{}

	otlpConnection struct {
		instance *trace.Instance
		client   *http.Client
		setting  otlpSetting
	}

	otlpSetting struct {
		Endpoint string
		Timeout  time.Duration
		Headers  map[string]string
		Service  string
	}
)

func init() {
	bamgoo.Register("otlp", &otlpDriver{})
}

func (d *otlpDriver) Connect(inst *trace.Instance) (trace.Connection, error) {
	setting := otlpSetting{
		Endpoint: "http://127.0.0.1:4318/v1/traces",
		Timeout:  5 * time.Second,
		Headers:  map[string]string{},
	}
	if inst != nil {
		if v, ok := getString(inst.Setting, "endpoint"); ok && v != "" {
			setting.Endpoint = v
		}
		if v, ok := getString(inst.Setting, "url"); ok && v != "" {
			setting.Endpoint = v
		}
		if v, ok := getDuration(inst.Setting, "timeout"); ok && v > 0 {
			setting.Timeout = v
		}
		if v, ok := getString(inst.Setting, "service"); ok && v != "" {
			setting.Service = v
		}
		if headers, ok := inst.Setting["headers"].(Map); ok {
			for k, v := range headers {
				if s, ok := v.(string); ok {
					setting.Headers[k] = s
				}
			}
		}
	}
	return &otlpConnection{instance: inst, setting: setting}, nil
}

func (c *otlpConnection) Open() error {
	c.client = &http.Client{Timeout: c.setting.Timeout}
	return nil
}

func (c *otlpConnection) Close() error { return nil }

func (c *otlpConnection) Write(spans ...trace.Span) error {
	if c.client == nil || len(spans) == 0 {
		return nil
	}
	payload := c.buildPayload(spans)
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.setting.Endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range c.setting.Headers {
		req.Header.Set(k, v)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("otlp export failed: status %d", resp.StatusCode)
	}
	return nil
}

func (c *otlpConnection) buildPayload(spans []trace.Span) Map {
	service := c.setting.Service
	if service == "" {
		service = c.instance.Name
	}
	fieldMap := trace.ResolveFields(c.instance.Setting, otlpDefaultFields())
	otlpSpans := make([]Any, 0, len(spans))
	for _, span := range spans {
		values := trace.SpanValues(span, c.instance.Name, c.instance.Config.Flag)
		serviceName := span.ServiceName
		if serviceName == "" {
			serviceName = service
		}
		attrs := []Any{
			kv("service.name", serviceName),
		}
		for k, v := range span.Attributes {
			attrs = append(attrs, kv(k, v))
		}
		for source, target := range fieldMap {
			if target == "" || target == "service.name" {
				continue
			}
			if val, ok := values[source]; ok {
				attrs = append(attrs, kv(target, val))
			}
		}
		statusCode := span.StatusCode
		if statusCode == "" {
			statusCode = "STATUS_CODE_OK"
			if span.Status == trace.StatusError {
				statusCode = "STATUS_CODE_ERROR"
			}
		}
		status := Map{"code": statusCode}
		if span.StatusMessage != "" {
			status["message"] = span.StatusMessage
		}
		startNano := span.StartTimeUnixNano
		if startNano <= 0 {
			startNano = span.StartMs * int64(time.Millisecond)
		}
		endNano := span.EndTimeUnixNano
		if endNano <= 0 {
			endNano = span.EndMs * int64(time.Millisecond)
		}
		if endNano <= 0 {
			endNano = span.Time.UnixNano()
		}
		if startNano <= 0 {
			startNano = endNano
		}
		otlpSpans = append(otlpSpans, Map{
			"traceId":           normalizeHex(span.TraceId, 32),
			"spanId":            normalizeHex(span.SpanId, 16),
			"parentSpanId":      normalizeHex(span.ParentSpanId, 16),
			"name":              span.Name,
			"kind":              kindCode(span.Kind),
			"startTimeUnixNano": strconv.FormatInt(startNano, 10),
			"endTimeUnixNano":   strconv.FormatInt(endNano, 10),
			"attributes":        attrs,
			"status":            status,
		})
	}
	return Map{
		"resourceSpans": []Any{
			Map{
				"resource": Map{
					"attributes": []Any{kv("service.name", service)},
				},
				"scopeSpans": []Any{
					Map{
						"scope": Map{"name": "bamgoo.trace"},
						"spans": otlpSpans,
					},
				},
			},
		},
	}
}

func otlpDefaultFields() map[string]string {
	return map[string]string{
		"trace_id":             "bamgoo.trace_id",
		"span_id":              "bamgoo.span_id",
		"parent_span_id":       "bamgoo.parent_span_id",
		"name":                 "bamgoo.name",
		"kind":                 "bamgoo.kind",
		"service_name":         "service.name",
		"target":               "bamgoo.target",
		"status":               "bamgoo.status",
		"status_code":          "bamgoo.status_code",
		"status_message":       "bamgoo.status_message",
		"duration_ms":          "bamgoo.duration_ms",
		"start_ms":             "bamgoo.start_ms",
		"end_ms":               "bamgoo.end_ms",
		"start_time_unix_nano": "bamgoo.start_time_unix_nano",
		"end_time_unix_nano":   "bamgoo.end_time_unix_nano",
		"timestamp":            "bamgoo.timestamp",
		"project":              "bamgoo.project",
		"profile":              "bamgoo.profile",
		"node":                 "bamgoo.node",
	}
}

func kv(key string, value Any) Map {
	entry := Map{"key": key}
	switch v := value.(type) {
	case bool:
		entry["value"] = Map{"boolValue": v}
	case int:
		entry["value"] = Map{"intValue": strconv.FormatInt(int64(v), 10)}
	case int64:
		entry["value"] = Map{"intValue": strconv.FormatInt(v, 10)}
	case float64:
		entry["value"] = Map{"doubleValue": v}
	case string:
		entry["value"] = Map{"stringValue": v}
	default:
		entry["value"] = Map{"stringValue": fmt.Sprintf("%v", value)}
	}
	return entry
}

func kindCode(kind string) int {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "internal":
		return 1
	case "server":
		return 2
	case "client":
		return 3
	case "producer":
		return 4
	case "consumer":
		return 5
	default:
		return 1
	}
}

func normalizeHex(v string, n int) string {
	v = strings.ToLower(strings.TrimSpace(v))
	v = strings.TrimPrefix(v, "0x")
	if v == "" {
		return strings.Repeat("0", n)
	}
	if len(v) > n {
		return v[len(v)-n:]
	}
	if len(v) < n {
		return strings.Repeat("0", n-len(v)) + v
	}
	return v
}

func getString(m Map, key string) (string, bool) {
	if m == nil {
		return "", false
	}
	v, ok := m[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func getDuration(m Map, key string) (time.Duration, bool) {
	if m == nil {
		return 0, false
	}
	val, ok := m[key]
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case time.Duration:
		return v, true
	case int:
		return time.Second * time.Duration(v), true
	case int64:
		return time.Second * time.Duration(v), true
	case float64:
		return time.Second * time.Duration(v), true
	case string:
		d, err := time.ParseDuration(v)
		if err == nil {
			return d, true
		}
	}
	return 0, false
}
