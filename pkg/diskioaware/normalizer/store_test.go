package normalizer

import (
	"testing"
)

var exec Normalize = func(in string) (string, error) {
	return in, nil
}

func TestNStore_Set(t *testing.T) {
	ns := NewnStore()
	type params struct {
		name string
		f    Normalize
	}
	tests := []struct {
		name     string
		p        *params
		expected bool
	}{
		{
			name: "Empty plugin name",
			p: &params{
				name: "",
				f:    nil,
			},
			expected: false,
		},
		{
			name: "Null exec func",
			p: &params{
				name: "p1",
				f:    nil,
			},
			expected: false,
		},
		{
			name: "Success",
			p: &params{
				name: "p1",
				f:    exec,
			},
			expected: true,
		},
		{
			name: "Duplicate plugin name",
			p: &params{
				name: "p1",
				f:    exec,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ns.Set(tt.p.name, tt.p.f)
			if (err == nil) != tt.expected {
				t.Errorf("case: %v failed got err=%v expected error=%v", tt.name, err, tt.expected)
			}
		})
	}
}

func TestNStore_Contains(t *testing.T) {
	ns := NewnStore()
	ns.Set("p1", exec)

	tests := []struct {
		name     string
		pName    string
		expected bool
	}{
		{
			name:     "Existing plugin",
			pName:    "p1",
			expected: true,
		},
		{
			name:     "Non-existing plugin",
			pName:    "p2",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ns.Contains(tt.pName)
			if got != tt.expected {
				t.Errorf("case: %v failed got=%v expected=%v", tt.name, got, tt.expected)
			}
		})
	}
}

// func TestnStore_Delete(t *testing.T) {
// 	ns := NewnStore()

// 	ns.Set("normalize1", NormalizeFunc(func(value int) int {
// 		return value * 2
// 	}))

// 	ns.Delete("normalize1")

// 	assert.False(t, ns.Contains("normalize1"))
// }

// func TestnStore_Get(t *testing.T) {
// 	ns := NewnStore()

// 	ns.Set("normalize1", NormalizeFunc(func(value int) int {
// 		return value * 2
// 	}))

// 	norm, err := ns.Get("normalize1")
// 	assert.NoError(t, err)
// 	assert.NotNil(t, norm)

// 	norm, err = ns.Get("normalize2")
// 	assert.Error(t, err)
// 	assert.Nil(t, norm)
// }
