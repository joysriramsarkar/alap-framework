// Package manager implements semantic versioning and package dependency resolution for NilPM.
package manager

import (
	"fmt"
	"strconv"
	"strings"
)

// Version represents a parsed Semantic Version (Major.Minor.Patch-Pre+Build).
type Version struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
}

// ParseVersion parses a semver string like "1.2.3".
func ParseVersion(vStr string) (Version, error) {
	vStr = strings.TrimPrefix(vStr, "v")
	parts := strings.Split(vStr, ".")
	if len(parts) < 3 {
		// Fill missing with 0
		for len(parts) < 3 {
			parts = append(parts, "0")
		}
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, err
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, err
	}
	patchParts := strings.Split(parts[2], "-")
	patch, err := strconv.Atoi(patchParts[0])
	if err != nil {
		return Version{}, err
	}

	pre := ""
	if len(patchParts) > 1 {
		pre = patchParts[1]
	}

	return Version{Major: major, Minor: minor, Patch: patch, Prerelease: pre}, nil
}

func (v Version) String() string {
	if v.Prerelease != "" {
		return fmt.Sprintf("%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.Prerelease)
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare returns -1 if v < other, 0 if v == other, 1 if v > other.
func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}
	return 0
}

// MatchesConstraint checks if version satisfies constraints like "^1.2.0" or "~1.2.3" or "*".
func (v Version) MatchesConstraint(constraint string) bool {
	constraint = strings.TrimSpace(constraint)
	if constraint == "*" || constraint == "latest" || constraint == "" {
		return true
	}

	if strings.HasPrefix(constraint, "^") {
		base, err := ParseVersion(strings.TrimPrefix(constraint, "^"))
		if err != nil {
			return false
		}
		// ^1.2.3 allows >=1.2.3 <2.0.0
		return v.Major == base.Major && (v.Minor > base.Minor || (v.Minor == base.Minor && v.Patch >= base.Patch))
	}

	if strings.HasPrefix(constraint, "~") {
		base, err := ParseVersion(strings.TrimPrefix(constraint, "~"))
		if err != nil {
			return false
		}
		// ~1.2.3 allows >=1.2.3 <1.3.0
		return v.Major == base.Major && v.Minor == base.Minor && v.Patch >= base.Patch
	}

	target, err := ParseVersion(constraint)
	if err != nil {
		return false
	}
	return v.Compare(target) == 0
}
