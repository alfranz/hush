package mathx

import "testing"

func TestClamp(t *testing.T) {
	for _, tt := range []struct{ v, want int }{{-5, 0}, {5, 5}, {50, 10}} {
		if got := Clamp(tt.v, 0, 10); got != tt.want {
			t.Errorf("Clamp(%d) = %d, want %d", tt.v, got, tt.want)
		}
	}
}

func TestMean(t *testing.T) {
	if got := Mean([]float64{1, 2, 3, 4}); got != 2.5 {
		t.Errorf("Mean = %v, want 2.5", got)
	}
	if got := Mean(nil); got != 0 {
		t.Errorf("Mean(nil) = %v, want 0", got)
	}
}
