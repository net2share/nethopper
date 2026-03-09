package actions

// OutputWriter defines the interface for action output.
type OutputWriter interface {
	Print(msg string)
	Printf(format string, args ...interface{})
	Println(args ...interface{})

	Info(msg string)
	Success(msg string)
	Warning(msg string)
	Error(msg string)

	Status(msg string)
	Step(current, total int, msg string)

	Box(title string, lines []string)
	KV(key, value string) string

	ShowInfo(cfg InfoConfig) error

	BeginProgress(title string)
	EndProgress()
	DismissProgress()
	IsProgressActive() bool
}

// InfoConfig configures an info display.
type InfoConfig struct {
	Title       string
	Description string
	Sections    []InfoSection
	CopyText    string
}

// InfoSection represents a section in the info view.
type InfoSection struct {
	Title string
	Rows  []InfoRow
}

// InfoRow represents a single row in an info section.
type InfoRow struct {
	Key   string
	Value string
}
