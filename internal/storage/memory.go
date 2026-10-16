package storage

import (
	"errors"
	"sort"
	"sync"
)

// ErrNotFound — общая (sentinel) ошибка «задача не найдена».
// Важно объявить её один раз: errors.Is сравнивает с конкретным значением,
// поэтому errors.Is(err, errors.New("not found")) всегда давало бы false.
var ErrNotFound = errors.New("not found")

type Task struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type MemoryStore struct {
	mu    sync.RWMutex
	auto  int64
	tasks map[int64]*Task
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		tasks: make(map[int64]*Task),
	}
}

func (s *MemoryStore) Create(title string) *Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.auto++
	t := &Task{ID: s.auto, Title: title, Done: false}
	s.tasks[t.ID] = t
	return copyTask(t)
}

func (s *MemoryStore) Get(id int64) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	if !ok {
		return nil, ErrNotFound
	}
	return copyTask(t), nil
}

// List возвращает задачи, отсортированные по ID (обход map в Go идёт в случайном порядке).
func (s *MemoryStore) List() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		out = append(out, copyTask(t))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// SetDone — для PATCH /tasks/{id}.
func (s *MemoryStore) SetDone(id int64, done bool) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok {
		return nil, ErrNotFound
	}
	t.Done = done
	return copyTask(t), nil
}

// Delete — для DELETE /tasks/{id}.
func (s *MemoryStore) Delete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(s.tasks, id)
	return nil
}

// copyTask — наружу отдаём копию, чтобы никто не менял данные в обход мьютекса.
func copyTask(t *Task) *Task {
	c := *t
	return &c
}
