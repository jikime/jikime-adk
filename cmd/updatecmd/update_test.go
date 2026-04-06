package updatecmd

import (
	"testing"
)

func TestCompareVersions_BasicComparison(t *testing.T) {
	// v1 < v2 -> -1
	result := compareVersions("1.0.0", "1.1.0")
	if result != -1 {
		t.Errorf("compareVersions(1.0.0, 1.1.0) = %d, want -1", result)
	}
}

func TestCompareVersions_EqualVersions(t *testing.T) {
	result := compareVersions("1.8.0", "1.8.0")
	if result != 0 {
		t.Errorf("compareVersions(1.8.0, 1.8.0) = %d, want 0", result)
	}
}

func TestCompareVersions_NewerVersion(t *testing.T) {
	// v1 > v2 -> 1
	result := compareVersions("2.0.0", "1.9.9")
	if result != 1 {
		t.Errorf("compareVersions(2.0.0, 1.9.9) = %d, want 1", result)
	}
}

func TestCompareVersions_DifferentLengths(t *testing.T) {
	// "1.8" vs "1.8.0" -> missing part treated as 0
	result := compareVersions("1.8", "1.8.0")
	if result != 0 {
		t.Errorf("compareVersions(1.8, 1.8.0) = %d, want 0", result)
	}
}

func TestCompareVersions_PreReleaseTag(t *testing.T) {
	// "1.8.0-rc1" vs "1.8.0" -> string comparison fallback on "0-rc1" vs "0"
	result := compareVersions("1.8.0-rc1", "1.8.0")
	// "0-rc1" cannot be parsed as int, falls back to strings.Compare("0-rc1", "0")
	// "0-rc1" > "0" lexicographically, so result should be > 0
	if result <= 0 {
		t.Errorf("compareVersions(1.8.0-rc1, 1.8.0) = %d, want > 0 (string comparison fallback)", result)
	}
}

func TestCompareVersions_MajorVersionDifference(t *testing.T) {
	result := compareVersions("1.0.0", "2.0.0")
	if result != -1 {
		t.Errorf("compareVersions(1.0.0, 2.0.0) = %d, want -1", result)
	}
}

func TestCompareVersions_PatchVersionDifference(t *testing.T) {
	result := compareVersions("1.0.1", "1.0.2")
	if result != -1 {
		t.Errorf("compareVersions(1.0.1, 1.0.2) = %d, want -1", result)
	}
}

func TestCompareVersions_SingleDigitVersions(t *testing.T) {
	result := compareVersions("1", "2")
	if result != -1 {
		t.Errorf("compareVersions(1, 2) = %d, want -1", result)
	}
}

func TestFindChecksumAssetURL_Found(t *testing.T) {
	assets := []GitHubAsset{
		{Name: "jikime-adk-linux-amd64", BrowserDownloadURL: "https://example.com/binary"},
		{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums.txt"},
	}

	url := findChecksumAssetURL(assets)
	if url != "https://example.com/checksums.txt" {
		t.Errorf("findChecksumAssetURL = %q, want checksums URL", url)
	}
}

func TestFindChecksumAssetURL_NotFound(t *testing.T) {
	assets := []GitHubAsset{
		{Name: "jikime-adk-linux-amd64", BrowserDownloadURL: "https://example.com/binary"},
	}

	url := findChecksumAssetURL(assets)
	if url != "" {
		t.Errorf("findChecksumAssetURL = %q, want empty string", url)
	}
}

func TestFindChecksumAssetURL_EmptyAssets(t *testing.T) {
	url := findChecksumAssetURL(nil)
	if url != "" {
		t.Errorf("findChecksumAssetURL(nil) = %q, want empty string", url)
	}
}
