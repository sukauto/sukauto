package controler

const (
	COMMAND        = "systemctl"
	GIT            = "git"
	PULL           = "pull"
	WORKDIR        = "WorkingDirectory"
	JournalCommand = "journalctl"
	STAT           = "status"
	STOP           = "stop"
	RUN            = "start"
	RESTART        = "restart"
	CFG_PATH       = "config.json"
	CmdEnable      = "enable"
	CmdDisable     = "disable"
	CmdShow        = "show"
	LogLimit       = 1024
)

// Modes
const (
	ModeUser = "--user"
	// journal
	ModeSystemUnit    = "-u"
	ModeUserUnit      = "--user-unit"
	ModeNoPages       = "--no-pager"
	ModeQuite         = "-q"
	ModeMergeJournals = "-m"
	ModeLimit         = "-n"
)

// Fields
const (
	FieldStatus = "SubState"
)

// Special states
const (
	StateUnknown = "unknown"
)

// Locations
const (
	LocationGlobal = "/etc/systemd/system"
	LocationUser   = "/.config/systemd/user" // prefix is $HOME
)
