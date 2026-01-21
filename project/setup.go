package project

// SetupAnswers holds values gathered during project initialization.
type SetupAnswers struct {
	ProjectName     string
	Locale          string
	UserName        string
	GitMode         string
	GitHubUser      string
	GitCommitLang   string
	CodeCommentLang string
	DocLang         string
	TagEnabled      bool   // TAG system enabled (SPEC-TAG-002)
	TagMode         string // TAG validation mode: warn, enforce, off
}
