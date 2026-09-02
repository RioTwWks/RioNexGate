package api

import (
	"sync"
	"time"
)

type ClientCommand struct {
	Type string `json:"type"`
}

type CommandManager struct {
	mu       sync.Mutex
	commands map[string][]ClientCommand
	waiters  map[string][]chan []ClientCommand
}

func NewCommandManager() *CommandManager {
	return &CommandManager{
		commands: make(map[string][]ClientCommand),
		waiters:  make(map[string][]chan []ClientCommand),
	}
}

func (m *CommandManager) Enqueue(deviceToken string, cmd ClientCommand) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if waiters, ok := m.waiters[deviceToken]; ok && len(waiters) > 0 {
		ch := waiters[0]
		m.waiters[deviceToken] = waiters[1:]
		ch <- []ClientCommand{cmd}
		close(ch)
		return
	}
	m.commands[deviceToken] = append(m.commands[deviceToken], cmd)
}

func (m *CommandManager) Poll(deviceToken string, timeout time.Duration) []ClientCommand {
	m.mu.Lock()
	if cmds, ok := m.commands[deviceToken]; ok && len(cmds) > 0 {
		result := cmds
		delete(m.commands, deviceToken)
		m.mu.Unlock()
		return result
	}

	ch := make(chan []ClientCommand, 1)
	m.waiters[deviceToken] = append(m.waiters[deviceToken], ch)
	m.mu.Unlock()

	select {
	case cmds := <-ch:
		return cmds
	case <-time.After(timeout):
		m.mu.Lock()
		waiters := m.waiters[deviceToken]
		for i, w := range waiters {
			if w == ch {
				m.waiters[deviceToken] = append(waiters[:i], waiters[i+1:]...)
				break
			}
		}
		m.mu.Unlock()
		return []ClientCommand{}
	}
}
