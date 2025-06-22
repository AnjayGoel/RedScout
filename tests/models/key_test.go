package models_test

import (
	"redscout/models"
	"testing"
)

func TestNewKey(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		delimiter string
		want      models.Key
		kp        *models.KeyParser
	}{
		{
			name:      "simple key without delimiter",
			input:     "mykey",
			delimiter: ":",
			want:      models.Key{"mykey"},
			kp:        models.NewKeyParser(":", nil),
		},
		{
			name:      "key with single delimiter",
			input:     "user:123",
			delimiter: ":",
			want:      models.Key{"user", "123"},
			kp:        models.NewKeyParser(":", nil),
		},
		{
			name:      "key with multiple delimiters",
			input:     "user:123:profile:settings",
			delimiter: ":",
			want:      models.Key{"user", "123", "profile", "settings"},
			kp:        models.NewKeyParser(":", nil),
		},
		{
			name:      "key with custom delimiter",
			input:     "user-123-profile",
			delimiter: "-",
			want:      models.Key{"user", "123", "profile"},
			kp:        models.NewKeyParser("-", nil),
		},
		{
			name:      "empty key",
			input:     "",
			delimiter: ":",
			want:      models.Key{""},
			kp:        models.NewKeyParser(":", nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.kp.NewKey(tt.input, false)
			if len(got) != len(tt.want) {
				t.Errorf("NewKey() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("NewKey()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestKeyIsA(t *testing.T) {
	tests := []struct {
		name   string
		key    models.Key
		prefix models.Key
		want   bool
		kp     *models.KeyParser
	}{
		{
			name:   "exact match",
			key:    models.Key{"user", "123"},
			prefix: models.Key{"user", "123"},
			want:   true,
			kp:     models.NewKeyParser(":", nil),
		},
		{
			name:   "prefix match",
			key:    models.Key{"user", "123", "profile", "settings"},
			prefix: models.Key{"user", "123"},
			want:   true,
			kp:     models.NewKeyParser(":", nil),
		},
		{
			name:   "no match",
			key:    models.Key{"user", "123"},
			prefix: models.Key{"user", "456"},
			want:   false,
			kp:     models.NewKeyParser(":", nil),
		},
		{
			name:   "prefix longer than key",
			key:    models.Key{"user"},
			prefix: models.Key{"user", "123"},
			want:   false,
			kp:     models.NewKeyParser(":", nil),
		},
		{
			name:   "empty prefix",
			key:    models.Key{"user", "123"},
			prefix: models.Key{},
			want:   true,
			kp:     models.NewKeyParser(":", nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.kp.IsA(tt.key, tt.prefix); got != tt.want {
				t.Errorf("Key.IsA() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyNamespace(t *testing.T) {
	tests := []struct {
		name    string
		key     models.Key
		prefix  models.Key
		want    string
		wantErr bool
		kp      *models.KeyParser
	}{
		{
			name:    "valid namespace",
			key:     models.Key{"user", "123", "profile", "settings"},
			prefix:  models.Key{"user", "123"},
			want:    "profile",
			wantErr: false,
			kp:      models.NewKeyParser(":", nil),
		},
		{
			name:    "empty prefix",
			key:     models.Key{"user", "123"},
			prefix:  models.Key{},
			want:    "user",
			wantErr: false,
			kp:      models.NewKeyParser(":", nil),
		},
		{
			name:    "exact prefix match",
			key:     models.Key{"user", "123"},
			prefix:  models.Key{"user", "123"},
			want:    "",
			wantErr: true,
			kp:      models.NewKeyParser(":", nil),
		},
		{
			name:    "invalid prefix",
			key:     models.Key{"user", "123"},
			prefix:  models.Key{"user", "456"},
			want:    "",
			wantErr: true,
			kp:      models.NewKeyParser(":", nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.kp.Namespace(tt.key, tt.prefix, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("Key.Namespace() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Key.Namespace() = %v, want %v", got, tt.want)
			}
		})
	}
}
