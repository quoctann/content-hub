package server

import (
	"context"
)

// HookFunc is a function that can be registered as a lifecycle hook
type HookFunc func() error

// Server defines the common interface for all server types in the system
type Server interface {
	Start() error
	Stop(ctx context.Context) error

	// Lifecycle hooks
	OnBeforeStart(fn HookFunc)
	OnAfterStart(fn HookFunc)
	OnBeforeStop(fn HookFunc)
	OnAfterStop(fn HookFunc)
}

// HookManager handles the registration and execution of lifecycle hooks
type HookManager struct {
	beforeStart []HookFunc
	afterStart  []HookFunc
	beforeStop  []HookFunc
	afterStop   []HookFunc
}

func (m *HookManager) OnBeforeStart(fn HookFunc) { m.beforeStart = append(m.beforeStart, fn) }
func (m *HookManager) OnAfterStart(fn HookFunc)  { m.afterStart = append(m.afterStart, fn) }
func (m *HookManager) OnBeforeStop(fn HookFunc)  { m.beforeStop = append(m.beforeStop, fn) }
func (m *HookManager) OnAfterStop(fn HookFunc)   { m.afterStop = append(m.afterStop, fn) }

func (m *HookManager) RunBeforeStart() error { return m.runHooks(m.beforeStart) }
func (m *HookManager) RunAfterStart() error  { return m.runHooks(m.afterStart) }
func (m *HookManager) RunBeforeStop() error  { return m.runHooks(m.beforeStop) }
func (m *HookManager) RunAfterStop() error   { return m.runHooks(m.afterStop) }

func (m *HookManager) runHooks(hooks []HookFunc) error {
	for _, fn := range hooks {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
