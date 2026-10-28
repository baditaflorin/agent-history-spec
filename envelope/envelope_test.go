package envelope_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/baditaflorin/agent-history-spec/envelope"
)

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestEventIDVectors(t *testing.T) {
	var v struct {
		Hash []struct {
			Name       string `json:"name"`
			SessionKey string `json:"session_key"`
			Line       int64  `json:"line"`
			Content    string `json:"content"`
			Want       string `json:"want"`
		} `json:"hash"`
		Native []struct {
			Name       string `json:"name"`
			Native     string `json:"native"`
			SessionKey string `json:"session_key"`
			Line       int64  `json:"line"`
			Content    string `json:"content"`
			Want       string `json:"want"`
		} `json:"native"`
		Scoped []struct {
			SessionKey string `json:"session_key"`
			Native     string `json:"native"`
			Want       string `json:"want"`
		} `json:"scoped"`
	}
	if err := json.Unmarshal(readFile(t, "../vectors/event-id.json"), &v); err != nil {
		t.Fatal(err)
	}
	if len(v.Hash) == 0 || len(v.Native) == 0 || len(v.Scoped) == 0 {
		t.Fatal("vector file has an empty section")
	}
	for _, c := range v.Hash {
		if got := envelope.HashID(c.SessionKey, c.Line, c.Content); got != c.Want {
			t.Errorf("%s: HashID = %s, want %s", c.Name, got, c.Want)
		}
	}
	for _, c := range v.Native {
		if got := envelope.ID(c.Native, c.SessionKey, c.Line, c.Content); got != c.Want {
			t.Errorf("%s: ID = %s, want %s", c.Name, got, c.Want)
		}
	}
	for _, c := range v.Scoped {
		if got := envelope.ScopedID(c.SessionKey, c.Native); got != c.Want {
			t.Errorf("ScopedID(%q, %q) = %s, want %s", c.SessionKey, c.Native, got, c.Want)
		}
	}
}

func readValid(t *testing.T) []envelope.Event {
	t.Helper()
	var events []envelope.Event
	err := envelope.ReadNDJSON(bytes.NewReader(readFile(t, "../vectors/valid.ndjson")), func(line int, e envelope.Event, derr error) error {
		if derr != nil {
			t.Fatalf("line %d: %v", line, derr)
		}
		events = append(events, e)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return events
}

func TestValidVectors(t *testing.T) {
	events := readValid(t)
	if len(events) != 7 {
		t.Fatalf("got %d events, want 7", len(events))
	}
	seen := map[string]bool{}
	for i := range events {
		if err := envelope.Validate(&events[i]); err != nil {
			t.Errorf("event %d (%s): %v", i, events[i].EventID, err)
		}
		if seen[events[i].EventID] {
			t.Errorf("duplicate event_id %s in vectors", events[i].EventID)
		}
		seen[events[i].EventID] = true
	}
	// exit code 0 must survive as "present", not vanish as a zero value
	if events[2].ExitCode == nil || *events[2].ExitCode != 0 {
		t.Errorf("exit_code 0 was lost: %v", events[2].ExitCode)
	}
	// an offset timestamp keeps its instant
	if want := time.Date(2026, 10, 21, 6, 15, 0, 500_000_000, time.UTC); !events[3].TS.Equal(want) {
		t.Errorf("ts = %v, want instant %v", events[3].TS, want)
	}
}

func TestRoundTrip(t *testing.T) {
	events := readValid(t)
	var buf bytes.Buffer
	if err := envelope.WriteNDJSON(&buf, events); err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(buf.String(), "\n"); got != len(events) {
		t.Fatalf("wrote %d lines for %d events", got, len(events))
	}
	var back []envelope.Event
	err := envelope.ReadNDJSON(&buf, func(_ int, e envelope.Event, derr error) error {
		if derr != nil {
			return derr
		}
		back = append(back, e)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(back) != len(events) {
		t.Fatalf("read back %d events, want %d", len(back), len(events))
	}
	for i := range events {
		a, _ := json.Marshal(events[i])
		b, _ := json.Marshal(back[i])
		if !bytes.Equal(a, b) {
			t.Errorf("event %d changed in a round trip:\n%s\n%s", i, a, b)
		}
	}
}

func TestInvalidVectors(t *testing.T) {
	var cases []struct {
		Name  string          `json:"name"`
		Field string          `json:"field"`
		Event json.RawMessage `json:"event"`
	}
	if err := json.Unmarshal(readFile(t, "../vectors/invalid.json"), &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		var e envelope.Event
		if err := json.Unmarshal(c.Event, &e); err != nil {
			t.Fatalf("%s: %v", c.Name, err)
		}
		err := envelope.Validate(&e)
		var fe *envelope.FieldError
		if !errors.As(err, &fe) {
			t.Errorf("%s: Validate returned %v, want a *FieldError", c.Name, err)
			continue
		}
		if fe.Field != c.Field {
			t.Errorf("%s: blamed %q, want %q (%s)", c.Name, fe.Field, c.Field, fe.Reason)
		}
	}
}

func baseEvent() envelope.Event {
	return envelope.Event{
		EventID: "e1", TS: time.Date(2026, 10, 20, 14, 0, 0, 0, time.UTC), Host: "mac14", Agent: "claude",
		SessionKey: "claude:mac14:abc", Role: envelope.RoleUser,
	}
}

func TestLimits(t *testing.T) {
	cases := []struct {
		name  string
		field string
		mut   func(*envelope.Event)
	}{
		{"content one byte over", "content", func(e *envelope.Event) { e.Content = strings.Repeat("a", envelope.MaxContentBytes+1) }},
		{"command one byte over", "command", func(e *envelope.Event) { e.Command = strings.Repeat("a", envelope.MaxContentBytes+1) }},
		{"multibyte content counted in bytes", "content", func(e *envelope.Event) { e.Content = strings.Repeat("é", envelope.MaxContentBytes/2+1) }},
		{"raw over its cap", "raw", func(e *envelope.Event) {
			e.Raw = json.RawMessage(`"` + strings.Repeat("a", envelope.MaxRawBytes) + `"`)
		}},
		{"raw that is not JSON", "raw", func(e *envelope.Event) { e.Raw = json.RawMessage(`{oops`) }},
		{"content that is not UTF-8", "content", func(e *envelope.Event) { e.Content = "a\xffb" }},
		{"too many edges", "edges", func(e *envelope.Event) {
			e.Edges = make([]envelope.Edge, envelope.MaxEdges+1)
		}},
		{"event_id with a control character", "event_id", func(e *envelope.Event) { e.EventID = "a\nb" }},
		{"event_id too long", "event_id", func(e *envelope.Event) { e.EventID = strings.Repeat("a", envelope.MaxIDBytes+1) }},
		{"title too long", "title", func(e *envelope.Event) { e.Title = strings.Repeat("a", envelope.MaxFieldBytes+1) }},
	}
	for _, c := range cases {
		e := baseEvent()
		c.mut(&e)
		err := envelope.Validate(&e)
		var fe *envelope.FieldError
		if !errors.As(err, &fe) || fe.Field != c.field {
			t.Errorf("%s: got %v, want a FieldError on %q", c.name, err, c.field)
		}
	}
	for _, ok := range []func(*envelope.Event){
		func(e *envelope.Event) { e.Content = strings.Repeat("a", envelope.MaxContentBytes) },
		func(e *envelope.Event) { e.Edges = make([]envelope.Edge, 0, envelope.MaxEdges) },
		func(e *envelope.Event) { e.SessionKey = "claude:mac14:a:b:c" }, // native ids may contain ':'
	} {
		e := baseEvent()
		ok(&e)
		if err := envelope.Validate(&e); err != nil {
			t.Errorf("a valid edge case was refused: %v", err)
		}
	}
}

func TestReadNDJSONBadLines(t *testing.T) {
	in := "\n{\"event_id\":\"a\"}\nnot json\n   \n{\"event_id\":\"b\"}\n"
	var lines []int
	var bads []int
	err := envelope.ReadNDJSON(strings.NewReader(in), func(line int, e envelope.Event, derr error) error {
		if derr != nil {
			bads = append(bads, line)
			return nil
		}
		lines = append(lines, line)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(lines, []int{2, 5}) || !reflect.DeepEqual(bads, []int{3}) {
		t.Fatalf("good lines %v bad lines %v", lines, bads)
	}
}

func TestReadNDJSONLineTooLong(t *testing.T) {
	long := `{"event_id":"` + strings.Repeat("a", envelope.MaxLineBytes) + `"}`
	in := "{\"event_id\":\"a\"}\n" + long + "\n{\"event_id\":\"c\"}\n"
	got := 0
	err := envelope.ReadNDJSON(strings.NewReader(in), func(int, envelope.Event, error) error { got++; return nil })
	if !errors.Is(err, envelope.ErrLineTooLong) {
		t.Fatalf("err = %v, want ErrLineTooLong", err)
	}
	if got != 1 {
		t.Fatalf("callback ran %d times before the long line, want 1", got)
	}
}

func TestReadNDJSONCallbackErrorStops(t *testing.T) {
	stop := errors.New("stop")
	n := 0
	err := envelope.ReadNDJSON(strings.NewReader("{}\n{}\n{}\n"), func(int, envelope.Event, error) error { n++; return stop })
	if !errors.Is(err, stop) || n != 1 {
		t.Fatalf("err = %v after %d calls", err, n)
	}
}

func TestWriteNDJSONDoesNotEscapeHTML(t *testing.T) {
	e := baseEvent()
	e.Command = "cat a > b && echo '<ok>'"
	var buf bytes.Buffer
	if err := envelope.WriteNDJSON(&buf, []envelope.Event{e}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), `<`) || strings.Contains(buf.String(), `&`) {
		t.Fatalf("HTML-escaped output: %s", buf.String())
	}
}

// The JSON Schema is for implementers outside Go. It must describe exactly the fields the struct
// carries and require exactly the ones Validate requires.
func TestSchemaMatchesStruct(t *testing.T) {
	var s struct {
		Required   []string                   `json:"required"`
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(readFile(t, "../schema/event.v1.schema.json"), &s); err != nil {
		t.Fatal(err)
	}
	var tags []string
	rt := reflect.TypeOf(envelope.Event{})
	for i := 0; i < rt.NumField(); i++ {
		name, _, _ := strings.Cut(rt.Field(i).Tag.Get("json"), ",")
		tags = append(tags, name)
	}
	var props []string
	for k := range s.Properties {
		props = append(props, k)
	}
	sort.Strings(tags)
	sort.Strings(props)
	if !reflect.DeepEqual(tags, props) {
		t.Errorf("struct fields %v\nschema properties %v", tags, props)
	}
	sort.Strings(s.Required)
	want := []string{"agent", "event_id", "host", "role", "session_key", "ts"}
	if !reflect.DeepEqual(s.Required, want) {
		t.Errorf("schema requires %v, want %v", s.Required, want)
	}
	// a struct with only the required fields must pass Validate, and removing any one must fail it
	for _, drop := range want {
		e := baseEvent()
		b, _ := json.Marshal(e)
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		delete(m, drop)
		b, _ = json.Marshal(m)
		var e2 envelope.Event
		_ = json.Unmarshal(b, &e2)
		if envelope.Validate(&e2) == nil {
			t.Errorf("Validate accepted an event without %s", drop)
		}
	}
}
