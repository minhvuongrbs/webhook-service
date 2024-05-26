package idgenerator

import "testing"

func TestNewId(t *testing.T) {
	type wantFunc func(string) bool

	tests := []struct {
		name string
		want wantFunc
	}{
		{
			name: "check format",
			want: func(s string) bool {
				if len(s) != 20 {
					return false
				}
				for _, c := range s {
					if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
						return false
					}
				}
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewId()
			t.Logf("new xid: %s", result)
			if !tt.want(result) {
				t.Errorf("case %s is fail, got: %s", tt.name, result)
			}
		})
	}
}
