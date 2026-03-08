package main

import (
	"fmt"
	"sync"
)

// Resetter интерфейс для типов-указателей, которые могут быть сброшены
// Метод Reset() должен быть определён на указателе (pointer receiver)
type Resetter interface {
	Reset()
}

// Pool универсальный пул объектов с generics
// T должен быть указателем на структуру, реализующую интерфейс Resetter
type Pool[T Resetter] struct {
	pool sync.Pool
}

// New конструктор для создания пула
// factory - функция-фабрика, создающая новый экземпляр типа T
func New[T Resetter](factory func() T) *Pool[T] {
	p := &Pool[T]{}
	p.pool.New = func() interface{} {
		return factory()
	}
	return p
}

// Get получение объекта из пула
// Возвращает указатель на тип T
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put возврат объекта в пул (сначала сбрасываем состояние методом Reset)
// Принимает объект типа T (указатель)
func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.pool.Put(obj)
}

// ============================================
// Пример структуры с методом Reset()
// ============================================

// Request структура для демонстрации работы пула
type Request struct {
	ID      int
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
}

// Reset сброс состояния объекта Request
// Метод определён на указателе (pointer receiver), так как sync.Pool работает с указателями
func (r *Request) Reset() {
	r.ID = 0
	r.Method = ""
	r.URL = ""
	// Очищаем map - создаём новую map для избежания утечек памяти
	r.Headers = make(map[string]string)
	// Обрезаем slice до нулевой длины (сохраняем ёмкость для повторного использования)
	r.Body = r.Body[:0]
}

// ============================================
// Демонстрация работы
// ============================================

func main() {
	// Создаём пул с фабрикой для создания новых Request
	// Тип T = *Request (указатель на Request)
	pool := New(func() *Request {
		return &Request{
			Headers: make(map[string]string),
		}
	})

	fmt.Println("=== Демонстрация работы Pool ===")
	fmt.Println()

	// Получаем первый объект из пула
	req1 := pool.Get()
	req1.ID = 1
	req1.Method = "GET"
	req1.URL = "https://example.com/api"
	req1.Headers["Content-Type"] = "application/json"
	req1.Body = append(req1.Body, []byte(`{"key": "value"}`)...)

	fmt.Printf("req1 после модификации:\n")
	fmt.Printf("  ID: %d, Method: %s, URL: %s\n", req1.ID, req1.Method, req1.URL)
	fmt.Printf("  Headers: %v\n", req1.Headers)
	fmt.Printf("  Body: %s\n", string(req1.Body))
	fmt.Println()

	// Возвращаем объект в пул (состояние будет сброшено методом Reset)
	pool.Put(req1)
	fmt.Println("req1 возвращён в пул (состояние сброшено)")
	fmt.Println()

	// Получаем объект повторно - должен быть сброшен
	req2 := pool.Get()
	fmt.Printf("req2 после получения из пула:\n")
	fmt.Printf("  ID: %d, Method: %s, URL: %s\n", req2.ID, req2.Method, req2.URL)
	fmt.Printf("  Headers: %v (len=%d)\n", req2.Headers, len(req2.Headers))
	fmt.Printf("  Body: %v (len=%d, cap=%d)\n", req2.Body, len(req2.Body), cap(req2.Body))
	fmt.Println()

	// Модифицируем и возвращаем снова
	req2.ID = 2
	req2.Method = "POST"
	req2.URL = "https://example.com/submit"
	req2.Headers["Authorization"] = "Bearer token123"
	req2.Body = append(req2.Body, []byte(`{"data": "test"}`)...)

	fmt.Printf("req2 после повторной модификации:\n")
	fmt.Printf("  ID: %d, Method: %s, URL: %s\n", req2.ID, req2.Method, req2.URL)
	fmt.Printf("  Headers: %v\n", req2.Headers)
	fmt.Printf("  Body: %s\n", string(req2.Body))
	fmt.Println()

	pool.Put(req2)
	fmt.Println("req2 возвращён в пул")
	fmt.Println()

	// Получаем третий раз - снова сброшен
	req3 := pool.Get()
	fmt.Printf("req3 после получения из пула:\n")
	fmt.Printf("  ID: %d, Method: %s, URL: %s\n", req3.ID, req3.Method, req3.URL)
	fmt.Println()

	// Демонстрация многопоточности
	fmt.Println("=== Демонстрация многопоточности ===")
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Получаем объект из пула
			req := pool.Get()
			req.ID = id
			req.Method = "GET"
			req.URL = fmt.Sprintf("https://example.com/%d", id)

			// Имитируем работу
			fmt.Printf("Goroutine %d использует Request с ID=%d\n", id, req.ID)

			// Возвращаем в пул
			pool.Put(req)
		}(i)
	}

	wg.Wait()
	fmt.Println()

	// Финальная проверка
	finalReq := pool.Get()
	fmt.Printf("Финальный Request из пула:\n")
	fmt.Printf("  ID: %d, Method: %s, URL: %s\n", finalReq.ID, finalReq.Method, finalReq.URL)
}
