package users

import (
	"fmt"
	"io/fs"
	"strconv"
	"strings"

	"github.com/spf13/afero"
)

// ParseQuotaString parses a quota string like "10M", "5G", "100m", "2g" to bytes.
// Returns 0 for unlimited (empty string or "0").
func ParseQuotaString(quota string) (int64, error) {
	quota = strings.TrimSpace(quota)
	if quota == "" || quota == "0" {
		return 0, nil
	}

	// Must have at least 2 characters (digit + unit)
	if len(quota) < 2 {
		return 0, fmt.Errorf("invalid quota format: %s", quota)
	}

	// Get unit (last character)
	unit := strings.ToUpper(string(quota[len(quota)-1]))
	valueStr := quota[:len(quota)-1]

	// Parse the numeric part
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid quota value: %s", quota)
	}

	if value < 0 {
		return 0, fmt.Errorf("quota cannot be negative: %s", quota)
	}

	// Convert to bytes based on unit
	var multiplier int64
	switch unit {
	case "M":
		multiplier = 1024 * 1024 // MB
	case "G":
		multiplier = 1024 * 1024 * 1024 // GB
	default:
		return 0, fmt.Errorf("invalid quota unit (must be M or G): %s", quota)
	}

	return value * multiplier, nil
}

// FormatQuotaBytes converts bytes to human-readable format (e.g., "10M", "5G")
func FormatQuotaBytes(bytes int64) string {
	if bytes == 0 {
		return "0"
	}

	const (
		GB = 1024 * 1024 * 1024
		MB = 1024 * 1024
	)

	if bytes >= GB && bytes%GB == 0 {
		return fmt.Sprintf("%dG", bytes/GB)
	}
	if bytes >= MB && bytes%MB == 0 {
		return fmt.Sprintf("%dM", bytes/MB)
	}

	// Default to MB for non-exact values
	return fmt.Sprintf("%dM", bytes/MB)
}

// CalculateUserUsage calculates the total size of all files in the user's filesystem
func CalculateUserUsage(afs afero.Fs) (int64, error) {
	var totalSize int64

	err := afero.Walk(afs, "/", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			// Skip files/directories we can't access
			return nil
		}

		if !info.IsDir() {
			totalSize += info.Size()
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return totalSize, nil
}

// CheckQuotaAvailable checks if there's enough quota available for additional storage
// Returns true if upload can proceed, false if quota would be exceeded
func CheckQuotaAvailable(currentUsage, quota, additionalSize int64) bool {
	// quota == 0 means unlimited
	if quota == 0 {
		return true
	}

	return currentUsage+additionalSize <= quota
}
