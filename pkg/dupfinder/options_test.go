package dupfinder

import "testing"

func TestOptionsValidation(t *testing.T) {
	opts := DefaultOptions()

	if err := opts.Validate(); err != nil {
		t.Errorf("expected default options to be valid, got: %v", err)
	}

	invalidOpts := DefaultOptions()
	invalidOpts.Workers = 0
	if err := invalidOpts.Validate(); err == nil {
		t.Errorf("expected options with 0 workers to be invalid")
	}

	invalidOpts = DefaultOptions()
	invalidOpts.Algorithm = "unknown"
	if err := invalidOpts.Validate(); err == nil {
		t.Errorf("expected options with unknown algorithm to be invalid")
	}
}
