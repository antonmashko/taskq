package taskq

import "sync"

type managerTask struct {
	task    Task
	manager *TaskManager
}

func (t *managerTask) Do() error {
	return nil
}

type TaskManager struct {
	taskQ *TaskQ
	lock  sync.RWMutex
	tasks map[int64]*managerTask
}

func NewTaskManger(taskQ *TaskQ) *TaskManager {
	return &TaskManager{
		taskQ: taskQ,
		tasks: make(map[int64]*managerTask),
	}
}

func (m *TaskManager) add(id int64, task Task) {
	m.lock.Lock()
	m.tasks[id] = &managerTask{task: task, manager: m}
	m.lock.Unlock()
}

func (m *TaskManager) Enqueue(task Task) int64 {
	id := m.taskQ.Enqueue(task)
	m.add(id, task)
	return id
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
