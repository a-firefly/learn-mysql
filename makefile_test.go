package main

import (
	"encoding/json"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

func TestMakeTransferRequest(t *testing.T) {
	output, err := exec.Command("make", "-n", "transfer", "1", "2", "100").CombinedOutput()
	if err != nil {
		t.Fatalf("make transfer: %v\n%s", err, output)
	}
	match := regexp.MustCompile(`-d '([^']+)'`).FindSubmatch(output)
	if len(match) != 2 {
		t.Fatalf("missing JSON payload in command:\n%s", output)
	}
	var input TransferInput
	if err := json.Unmarshal(match[1], &input); err != nil {
		t.Fatalf("decode transfer request: %v", err)
	}
	if input != (TransferInput{FromID: 1, ToID: 2, Cents: 100}) {
		t.Fatalf("unexpected transfer input: %+v", input)
	}
	if strings.Count(string(output), "curl ") != 1 {
		t.Fatalf("transfer arguments triggered extra requests:\n%s", output)
	}
}
