package xbt

import "testing"

const (
	xbtToolsUrl = "http://127.0.0.1:3000" //xbt-tools
)

func TestXbtToolsPostCall(t *testing.T) {
	tw := NewClient(xbtToolsUrl, true, symbol, currencyDecimal)

	body := map[string]interface{}{
		"private" : "xxx",
		"to" : "xB029b2bc3302ddaF67953bF98F0C88EEFde7e5e9D",
		"amount" : 10,
	}

	if r, err := tw.PostCall("/transaction/send", body); err != nil {
		t.Errorf("Post Call Result failed: %v\n", err)
	} else {
		PrintJsonLog(t, r.String())
	}
}
