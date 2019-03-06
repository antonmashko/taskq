package taskq

import "sync"

type TaskManager struct {
	taskQ *TaskQ
	lock  sync.RWMutex
	tasks map[int64]*itask
}

func NewTaskManger(taskQ *TaskQ) *TaskManager {
	return &TaskManager{
		taskQ: taskQ,
		tasks: make(map[int64]*itask),
	}
}

func (m *TaskManager) Enqueue(task Task) int64 {
	it := m.taskQ.enqueue(task)
	if it == nil {
		return -1
	}
	m.lock.Lock()
	m.tasks[it.id] = it
	m.lock.Unlock()
	return it.id
}

func (m *TaskManager) Task(id int64) Task {
	m.lock.RLock()
	result := m.tasks[id]
	if result == nil {
		m.lock.RUnlock()
		return nil
	}
	m.lock.RUnlock()
	return result
}

func (m *TaskManager) Start() error {
	return m.taskQ.Start()
}

func (m *TaskManager) Close() error {
	return m.taskQ.Close()
}
