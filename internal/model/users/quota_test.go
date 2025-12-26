package users

import "testing"

func TestParseQuotaString(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"10M", 10 * 1024 * 1024, false},
		{"5G", 5 * 1024 * 1024 * 1024, false},
		{"100m", 100 * 1024 * 1024, false},
		{"2g", 2 * 1024 * 1024 * 1024, false},
		{"0", 0, false},
		{"", 0, false},
		{" 10M ", 10 * 1024 * 1024, false},
		{"abc", 0, true},
		{"10K", 0, true},
		{"-5G", 0, true},
		{"M", 0, true},
		{"10", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseQuotaString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQuotaString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ParseQuotaString(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatQuotaBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{10 * 1024 * 1024, "10M"},
		{5 * 1024 * 1024 * 1024, "5G"},
		{1536 * 1024 * 1024, "1536M"}, // 1.5GB, not exact
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := FormatQuotaBytes(tt.input)
			if got != tt.expected {
				t.Errorf("FormatQuotaBytes(%d) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCheckQuotaAvailable(t *testing.T) {
	tests := []struct {
		name           string
		currentUsage   int64
		quota          int64
		additionalSize int64
		expected       bool
	}{
		{"unlimited quota", 5 * 1024 * 1024, 0, 10 * 1024 * 1024, true},
		{"within quota", 5 * 1024 * 1024, 10 * 1024 * 1024, 3 * 1024 * 1024, true},
		{"exactly at quota", 7 * 1024 * 1024, 10 * 1024 * 1024, 3 * 1024 * 1024, true},
		{"exceeds quota", 8 * 1024 * 1024, 10 * 1024 * 1024, 3 * 1024 * 1024, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckQuotaAvailable(tt.currentUsage, tt.quota, tt.additionalSize)
			if got != tt.expected {
				t.Errorf("CheckQuotaAvailable(%d, %d, %d) = %v, want %v",
					tt.currentUsage, tt.quota, tt.additionalSize, got, tt.expected)
			}
		})
	}
}
