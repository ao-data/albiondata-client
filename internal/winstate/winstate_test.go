package winstate

import "testing"

func withTempConfigDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	original := userConfigDir
	userConfigDir = func() (string, error) { return dir, nil }
	t.Cleanup(func() { userConfigDir = original })
}

func TestLoad_NothingSavedYet(t *testing.T) {
	withTempConfigDir(t)

	_, ok := Load()
	if ok {
		t.Fatalf("expected Load to report nothing saved yet")
	}
}

func TestSaveThenLoad_RoundTrips(t *testing.T) {
	withTempConfigDir(t)

	want := Bounds{X: 120, Y: 80, Width: 1024, Height: 768}
	if err := Save(want); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, ok := Load()
	if !ok {
		t.Fatalf("expected Load to find the saved bounds")
	}
	if got != want {
		t.Fatalf("Load returned %+v, want %+v", got, want)
	}
}

func TestSave_OverwritesPreviousBounds(t *testing.T) {
	withTempConfigDir(t)

	if err := Save(Bounds{X: 0, Y: 0, Width: 900, Height: 600}); err != nil {
		t.Fatalf("first Save failed: %v", err)
	}
	want := Bounds{X: 50, Y: 60, Width: 1200, Height: 900}
	if err := Save(want); err != nil {
		t.Fatalf("second Save failed: %v", err)
	}

	got, ok := Load()
	if !ok {
		t.Fatalf("expected Load to find the saved bounds")
	}
	if got != want {
		t.Fatalf("Load returned %+v, want %+v", got, want)
	}
}

func TestLoad_RejectsZeroSizedBounds(t *testing.T) {
	withTempConfigDir(t)

	if err := Save(Bounds{X: 10, Y: 10, Width: 0, Height: 0}); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	_, ok := Load()
	if ok {
		t.Fatalf("expected Load to reject zero-sized bounds")
	}
}
