package version

// buildVersion is injected at build time via -ldflags "-X 'jikime-adk/version.buildVersion=<value>'".
var buildVersion string

const fallbackVersion = "0.6.0"

func String() string {
	if buildVersion != "" {
		return buildVersion
	}
	return fallbackVersion
}
