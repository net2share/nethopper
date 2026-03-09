// Package actions provides the unified action system for nethopper CLI and menu.
package actions

import "context"

// InputType defines the type of input field.
type InputType int

const (
	InputTypeText InputType = iota
	InputTypePassword
	InputTypeSelect
	InputTypeNumber
	InputTypeBool
)

// SelectOption defines an option for select inputs.
type SelectOption struct {
	Label       string
	Value       string
	Description string
}

// InputField defines an input field for an action.
type InputField struct {
	Name            string
	Label           string
	Description     string
	Type            InputType
	Required        bool
	Default         string
	Placeholder     string
	Options         []SelectOption
	OptionsFunc     func(ctx *Context) []SelectOption
	ShortFlag       rune
	ShowIf          func(ctx *Context) bool
	Validate        func(value string) error
	DefaultFunc     func(ctx *Context) string
	InteractiveOnly bool
	DescriptionFunc func(ctx *Context) string
}

// ConfirmConfig defines confirmation settings for an action.
type ConfirmConfig struct {
	Message    string
	DefaultNo  bool
	ForceFlag  string
}

// ArgsSpec defines the positional arguments for an action.
type ArgsSpec struct {
	Name        string
	Description string
	Required    bool
	PickerFunc  func(ctx *Context) (string, error)
}

// Handler is the function signature for action handlers.
type Handler func(ctx *Context) error

// Action defines a command/menu action.
type Action struct {
	ID               string
	Parent           string
	Use              string
	Short            string
	Long             string
	MenuLabel        string
	Args             *ArgsSpec
	Inputs           []InputField
	Confirm          *ConfirmConfig
	Handler          Handler
	RequiresRoot     bool
	Hidden           bool
	IsSubmenu        bool
}

// Context provides the execution context for action handlers.
type Context struct {
	Ctx           context.Context
	Args          []string
	Values        map[string]interface{}
	Output        OutputWriter
	IsInteractive bool
}

// GetString returns a string value from the context.
func (c *Context) GetString(key string) string {
	if v, ok := c.Values[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetInt returns an integer value from the context.
func (c *Context) GetInt(key string) int {
	if v, ok := c.Values[key]; ok {
		switch i := v.(type) {
		case int:
			return i
		case int64:
			return int(i)
		case float64:
			return int(i)
		}
	}
	return 0
}

// GetBool returns a boolean value from the context.
func (c *Context) GetBool(key string) bool {
	if v, ok := c.Values[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// GetArg returns the positional argument at the given index.
func (c *Context) GetArg(index int) string {
	if index >= 0 && index < len(c.Args) {
		return c.Args[index]
	}
	return ""
}

// Set sets a value in the context.
func (c *Context) Set(key string, value interface{}) {
	if c.Values == nil {
		c.Values = make(map[string]interface{})
	}
	c.Values[key] = value
}
