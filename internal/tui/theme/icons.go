package theme

var Icons = struct {
	Connected    string
	Disconnected string
	Running      string
	Success      string
	Error        string
	Warning      string
	Info         string
	Loading      string
	Close        string

	Modified string
	Saved    string
	Check    string
}{
	Connected:    "[OK]",
	Disconnected: "[--]",
	Running:      "[..]",
	Success:      "[OK]",
	Error:        "[ERR]",
	Warning:      "[!]",
	Info:         "[i]",
	Loading:      "[...]",
	Close:        "[X]",

	Modified: "[*]",
	Saved:    "[=]",
	Check:    "[x]",
}
