package composechk

import "testing"

func TestCheckBasic(t *testing.T) {
	ok := Check("t.yml", "services:\n  web:\n    image: nginx:1.27\n  db:\n    image: postgres:16\n")
	for _, f := range ok.Findings {
		if f.Level == "error" {
			t.Fatalf("unexpected error: %+v", ok.Findings)
		}
	}
	bad := Check("t.yml", "services:\n  web:\n    command: ./run\n")
	found := false
	for _, f := range bad.Findings {
		if f.Level == "error" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected error for missing image/build: %+v", bad.Findings)
	}
}
