package apicompat

import (
	"encoding/json"
	"strings"
	"testing"
)

// Regression: newer Codex clients send function_call_output.output as a
// content-part array. Before the fix the whole input array failed to decode
// ("cannot unmarshal array into Go struct field ResponsesInputItem.output of
// type string"), which reached clients as a 502.
func TestResponsesInputItemUnmarshalArrayOutput(t *testing.T) {
	raw := []byte(`[
		{"type":"function_call","call_id":"call_1","name":"shell","arguments":"{}"},
		{"type":"function_call_output","call_id":"call_1","output":[{"type":"input_text","text":"hello from tool"}]}
	]`)

	var items []ResponsesInputItem
	if err := json.Unmarshal(raw, &items); err != nil {
		t.Fatalf("unmarshal array-form output: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("want 2 items, got %d", len(items))
	}
	if got := items[1].Type; got != "function_call_output" {
		t.Fatalf("want function_call_output, got %q", got)
	}
	if len(items[1].outputRaw) == 0 {
		t.Fatal("outputRaw should preserve the array form")
	}
}

func TestResponsesInputItemUnmarshalStringOutputUnchanged(t *testing.T) {
	var item ResponsesInputItem
	if err := json.Unmarshal([]byte(`{"type":"function_call_output","call_id":"c1","output":"plain"}`), &item); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if item.Output != "plain" {
		t.Fatalf("want Output=plain, got %q", item.Output)
	}
	if len(item.outputRaw) != 0 {
		t.Fatal("string output must not populate outputRaw")
	}
}

func TestConvertResponsesInputToAnthropicArrayOutput(t *testing.T) {
	raw := json.RawMessage(`[
		{"type":"message","role":"user","content":"run it"},
		{"type":"function_call","call_id":"call_1","name":"shell","arguments":"{}"},
		{"type":"function_call_output","call_id":"call_1","output":[{"type":"input_text","text":"tool said hi"}]}
	]`)

	_, msgs, err := convertResponsesInputToAnthropic("", raw)
	if err != nil {
		t.Fatalf("convert must not fail on array-form output: %v", err)
	}

	var found bool
	for _, m := range msgs {
		var blocks []AnthropicContentBlock
		if json.Unmarshal(m.Content, &blocks) != nil {
			continue
		}
		for _, b := range blocks {
			if b.Type != "tool_result" {
				continue
			}
			found = true
			// The array must be rehydrated into typed blocks, not stringified.
			var inner []AnthropicContentBlock
			if err := json.Unmarshal(b.Content, &inner); err != nil {
				t.Fatalf("tool_result content should be a block array, got %s", b.Content)
			}
			if len(inner) != 1 || inner[0].Type != "text" || inner[0].Text != "tool said hi" {
				t.Fatalf("unexpected rehydrated content: %s", b.Content)
			}
		}
	}
	if !found {
		t.Fatal("no tool_result block produced")
	}
}

func TestConvertResponsesInputToAnthropicArrayOutputWithImage(t *testing.T) {
	raw := json.RawMessage(`[
		{"type":"function_call","call_id":"c1","name":"screenshot","arguments":"{}"},
		{"type":"function_call_output","call_id":"c1","output":[
			{"type":"input_text","text":"see below"},
			{"type":"input_image","image_url":"data:image/png;base64,aGk="}
		]}
	]`)

	_, msgs, err := convertResponsesInputToAnthropic("", raw)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}

	var texts, images int
	for _, m := range msgs {
		var blocks []AnthropicContentBlock
		if json.Unmarshal(m.Content, &blocks) != nil {
			continue
		}
		for _, b := range blocks {
			if b.Type != "tool_result" {
				continue
			}
			var inner []AnthropicContentBlock
			if json.Unmarshal(b.Content, &inner) != nil {
				continue
			}
			for _, ib := range inner {
				switch ib.Type {
				case "text":
					texts++
				case "image":
					images++
				}
			}
		}
	}
	if texts != 1 || images != 1 {
		t.Fatalf("want 1 text + 1 image block, got %d text / %d image", texts, images)
	}
}

// Empty string output keeps the historical "(empty)" placeholder.
func TestConvertResponsesInputToAnthropicEmptyStringOutput(t *testing.T) {
	raw := json.RawMessage(`[
		{"type":"function_call","call_id":"c1","name":"noop","arguments":"{}"},
		{"type":"function_call_output","call_id":"c1","output":""}
	]`)

	_, msgs, err := convertResponsesInputToAnthropic("", raw)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if !strings.Contains(marshalAll(t, msgs), "(empty)") {
		t.Fatal(`want "(empty)" placeholder for empty output`)
	}
}

// Round-trip guard: continuation trimming and todo-guard injection decode the
// input array and marshal it back. An array-form output must survive as an
// array, not become a JSON string.
func TestResponsesInputItemMarshalPreservesArrayOutput(t *testing.T) {
	raw := []byte(`[{"type":"function_call_output","call_id":"c1","output":[{"type":"input_text","text":"hi"}]}]`)

	var items []ResponsesInputItem
	if err := json.Unmarshal(raw, &items); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	out, err := json.Marshal(items)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var probe []struct {
		Output json.RawMessage `json:"output"`
	}
	if err := json.Unmarshal(out, &probe); err != nil {
		t.Fatalf("re-unmarshal: %v", err)
	}
	var parts []ResponsesContentPart
	if err := json.Unmarshal(probe[0].Output, &parts); err != nil {
		t.Fatalf("output should round-trip as an array, got %s", probe[0].Output)
	}
	if len(parts) != 1 || parts[0].Text != "hi" {
		t.Fatalf("unexpected round-tripped parts: %s", probe[0].Output)
	}
}

func TestResponsesInputItemMarshalStringOutputUnchanged(t *testing.T) {
	item := ResponsesInputItem{Type: "function_call_output", CallID: "c1", Output: "plain"}
	out, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	const want = `{"type":"function_call_output","call_id":"c1","output":"plain"}`
	if string(out) != want {
		t.Fatalf("serialisation changed:\n got %s\nwant %s", out, want)
	}
}

func marshalAll(t *testing.T, msgs []AnthropicMessage) string {
	t.Helper()
	b, err := json.Marshal(msgs)
	if err != nil {
		t.Fatalf("marshal messages: %v", err)
	}
	return string(b)
}
