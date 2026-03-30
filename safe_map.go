package test_task

import "sync"

// Автор: Барыкин Евгений (evgeniy.barykin@gmail.com)
// SafeMap — map, защищенная мьютексом, с тестовыми счетчиками.
type SafeMap struct {
	mu          sync.Mutex
	data        map[int]int
	accessCount int
	insertCount int
}

// NewSafeMap создает пустую SafeMap.
func NewSafeMap() *SafeMap {
	return &SafeMap{
		data: make(map[int]int),
	}
}

// Get возвращает значение по ключу и увеличивает счетчик обращений.
func (m *SafeMap) Get(key int) (int, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.accessCount++
	v, ok := m.data[key]
	return v, ok
}

// Set записывает значение по ключу и увеличивает счетчик добавлений для новых ключей.
func (m *SafeMap) Set(key int, value int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.data[key]; !exists {
		m.insertCount++
	}
	m.data[key] = value
}

// Stats возвращает текущие значения счетчиков обращений и добавлений.
func (m *SafeMap) Stats() (accessCount int, insertCount int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.accessCount, m.insertCount
}

// Snapshot возвращает копию данных для безопасных проверок только на чтение.
func (m *SafeMap) Snapshot() map[int]int {
	m.mu.Lock()
	defer m.mu.Unlock()

	cp := make(map[int]int, len(m.data))
	for k, v := range m.data {
		cp[k] = v
	}
	return cp
}
