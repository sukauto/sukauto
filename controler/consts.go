package controler

const (
	COMMAND    = "systemctl"
	STAT       = "status"
	STOP       = "stop"
	RUN        = "start"
	RESTART    = "restart"
	CFG_PATH   = "config.json"
	CmdEnable  = "enable"
	CmdDisable = "disable"
	CmdShow    = "show"
)

// Modes
const (
	ModeUser = "--user"
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
