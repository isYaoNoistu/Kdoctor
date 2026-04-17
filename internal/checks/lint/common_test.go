package lint

import "testing"

func TestParseListeners(t *testing.T) {
	listeners, err := parseListeners("INTERNAL://192.168.1.10:9192,EXTERNAL://0.0.0.0:9292,CONTROLLER://192.168.1.10:9193")
	if err != nil {
		t.Fatalf("parse listeners: %v", err)
	}
	if listeners["INTERNAL"].Port != 9192 {
		t.Fatalf("expected INTERNAL port 9192, got %d", listeners["INTERNAL"].Port)
	}
	if listeners["EXTERNAL"].Host != "0.0.0.0" {
		t.Fatalf("expected EXTERNAL host 0.0.0.0, got %q", listeners["EXTERNAL"].Host)
	}
}

func TestParseVoters(t *testing.T) {
	voters, err := parseVoters("1@192.168.1.10:9193,2@192.168.1.10:9195,3@192.168.1.10:9197")
	if err != nil {
		t.Fatalf("parse voters: %v", err)
	}
	if len(voters) != 3 {
		t.Fatalf("expected 3 voters, got %d", len(voters))
	}
	if voters[2] != "192.168.1.10:9195" {
		t.Fatalf("unexpected voter 2 value %q", voters[2])
	}
}
