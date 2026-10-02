package passwords

import "testing"

func TestEstimateStrength(t *testing.T) {
	cases := []struct {
		name     string
		password string
		weak     bool // expected to be at/below the weak threshold
	}{
		{"empty", "", true},
		{"short", "aB3$", true},
		{"common", "password", true},
		{"common mixed case", "Password", true},
		{"only lower long", "abcdefghijkl", true},
		{"decent", "Sunflower12", false},
		{"strong", "Tr0ub4dour&3xplain", false},
		{"very strong", "9xK#mQ2!vL7@wP4zB8%r", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := EstimateStrength(c.password)
			if got < 0 || got > 4 {
				t.Fatalf("score out of range: %d", got)
			}
			weak := got <= WeakStrengthThreshold
			if weak != c.weak {
				t.Errorf("EstimateStrength(%q) = %d (weak=%v), want weak=%v", c.password, got, weak, c.weak)
			}
		})
	}
}
