package version

// buildVersion is injected at build time via -ldflags "-X 'jikime-adk/version.buildVersion=<value>'".
var buildVersion string

const fallbackVersion = "1.7.17"

func String() string {
	if buildVersion != "" {
		return buildVersion
	}
	return fallbackVersion
}
