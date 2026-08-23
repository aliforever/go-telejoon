package telejoon

import "strings"

// Command represents a parsed Telegram command with its arguments
type Command struct {
	Name    string   // Command name without the leading /
	Args    []string // Arguments after the command, split by whitespace
	RawArgs string   // Raw argument string (everything after the command)
	BotName string   // Bot username if specified (e.g., /command@botname)
}

// ParseCommand parses a Telegram command from message text.
// Returns nil if the text is not a command (doesn't start with /).
//
// Examples:
//   - "/start" → Name: "start", Args: [], RawArgs: ""
//   - "/help arg1 arg2" → Name: "help", Args: ["arg1", "arg2"], RawArgs: "arg1 arg2"
//   - "/command@botname arg" → Name: "command", BotName: "botname", Args: ["arg"]
func ParseCommand(text string) *Command {
	if !strings.HasPrefix(text, "/") {
		return nil
	}

	parts := strings.SplitN(text, " ", 2)
	cmdPart := strings.TrimPrefix(parts[0], "/")

	// Handle @botname suffix
	var botName string
	if idx := strings.Index(cmdPart, "@"); idx != -1 {
		botName = cmdPart[idx+1:]
		cmdPart = cmdPart[:idx]
	}

	if cmdPart == "" {
		return nil
	}

	cmd := &Command{
		Name:    cmdPart,
		BotName: botName,
	}

	if len(parts) > 1 {
		cmd.RawArgs = parts[1]
		cmd.Args = strings.Fields(parts[1])
	}

	return cmd
}

// IsCommand checks if text starts with a command prefix (/)
func IsCommand(text string) bool {
	return strings.HasPrefix(text, "/")
}

// Arg returns the argument at the given index, or empty string if out of range
func (c *Command) Arg(index int) string {
	if index < 0 || index >= len(c.Args) {
		return ""
	}
	return c.Args[index]
}

// ArgOr returns the argument at the given index, or the default value if out of range
func (c *Command) ArgOr(index int, defaultValue string) string {
	if index < 0 || index >= len(c.Args) {
		return defaultValue
	}
	return c.Args[index]
}

// HasArgs returns true if the command has any arguments
func (c *Command) HasArgs() bool {
	return len(c.Args) > 0
}

// ArgCount returns the number of arguments
func (c *Command) ArgCount() int {
	return len(c.Args)
}
