package actions

import (
	"sort"
	"strings"
	"sync"
)

var (
	registry = make(map[string]*Action)
	mu       sync.RWMutex
)

// Register adds an action to the registry.
func Register(action *Action) {
	mu.Lock()
	defer mu.Unlock()
	registry[action.ID] = action
}

// Get retrieves an action by ID.
func Get(id string) *Action {
	mu.RLock()
	defer mu.RUnlock()
	return registry[id]
}

// All returns all registered actions sorted by ID.
func All() []*Action {
	mu.RLock()
	defer mu.RUnlock()

	actions := make([]*Action, 0, len(registry))
	for _, action := range registry {
		actions = append(actions, action)
	}
	sort.Slice(actions, func(i, j int) bool {
		return actions[i].ID < actions[j].ID
	})
	return actions
}

// ByParent returns all actions with the given parent.
func ByParent(parentID string) []*Action {
	mu.RLock()
	defer mu.RUnlock()

	var actions []*Action
	for _, action := range registry {
		if action.Parent == parentID {
			actions = append(actions, action)
		}
	}
	sort.Slice(actions, func(i, j int) bool {
		return actions[i].ID < actions[j].ID
	})
	return actions
}

// TopLevel returns all top-level actions (no parent).
func TopLevel() []*Action {
	return ByParent("")
}

// GetChildren returns immediate children of an action.
func GetChildren(actionID string) []*Action {
	return ByParent(actionID)
}

// GetCommandName returns the command name portion of an action ID.
func GetCommandName(actionID string) string {
	parts := strings.Split(actionID, ".")
	return parts[len(parts)-1]
}

// SetHandler sets the handler for an action.
func SetHandler(actionID string, handler Handler) {
	mu.Lock()
	defer mu.Unlock()
	if action, ok := registry[actionID]; ok {
		action.Handler = handler
	}
}

// Clear clears the registry (useful for testing).
func Clear() {
	mu.Lock()
	defer mu.Unlock()
	registry = make(map[string]*Action)
}
