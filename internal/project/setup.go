package project

// SetupAnswers holds values gathered during project initialization.
type SetupAnswers struct {
	ProjectName     string
	Locale          string
	UserName        string
	Honorific       string // User's preferred honorific (e.g., "sir", "님", "さん")
	TonePreset      string // Tone preset: friendly, professional, casual, mentor
	GitMode         string
	GitHubUser      string
	GitCommitLang   string
	CodeCommentLang string
	DocLang         string
	TagEnabled      bool   // TAG system enabled (SPEC-TAG-002)
	TagMode         string // TAG validation mode: warn, enforce, off
}
