package pipeline

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Настройка логгера, чтобы писать в консоль с уровнем Info
func init() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))
}

// Интервал очистки кольцевого буфера
const bufferDrainInterval time.Duration = 10 * time.Second

// Размер кольцевого буфера
const bufferSize int = 1000

// Узел связного списка
type Node struct {
	value int
	next  *Node
}

// Структура кольцевого буфера
type RingBuffer struct {
	head *Node
	tail *Node
	pos  int
	size int
	m    sync.Mutex
}

// Конструктор для буфера
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{size: size}
}

// Добавление элемента в кольцевой буфер
func (r *RingBuffer) Push(value int) {
	r.m.Lock()
	defer r.m.Unlock()
	newNode := &Node{value: value}

	if r.pos == 0 {
		r.head = newNode
		r.tail = newNode
		r.tail.next = r.head
		r.pos++
		return
	}

	if r.pos < r.size {
		r.tail.next = newNode
		r.tail = newNode
		r.tail.next = r.head
		r.pos++
		return
	}

	r.head.value = value
	r.tail = r.head
	r.head = r.head.next
}

// Получение всех элементов кольцевого буфера и их удаление
func (r *RingBuffer) Get() []int {
	r.m.Lock()
	defer r.m.Unlock()

	if r.pos == 0 {
		fmt.Println("Буфер пуст")
		return nil
	}

	output := make([]int, 0, r.pos)
	curr := r.head

	for i := 0; i < r.pos; i++ {
		output = append(output, curr.value)
		curr = curr.next
	}

	r.head = nil
	r.tail = nil
	r.pos = 0

	return output
}

// Считывание данных с консоли
func Read(nextStage chan<- int, done chan struct{}) {
	fmt.Println("Для завершения программы напишите 'exit'")
	scanner := bufio.NewScanner(os.Stdin)
	var data string

	for scanner.Scan() {
		data = scanner.Text()
		if strings.EqualFold(data, "exit") {
			slog.Info("Программа завершила работу по команде пользователя")
			close(done)
			return
		}

		n, err := strconv.Atoi(data)
		if err != nil {
			fmt.Println("Доступен ввод только целых чисел")
			slog.Warn("Попытка ввода нечислового значения", "input", data)
			continue
		}

		slog.Debug("Пользователь ввел число", "value", n)
		nextStage <- n
	}

	if err := scanner.Err(); err != nil {
		slog.Error("Ошибка чтения с консоли", "error", err)
	}
}

// Стадия фильтрации отрицательных чисел
func NegativeFilterStage(prevStage <-chan int, done <-chan struct{}) <-chan int {
	nextStage := make(chan int)

	go func() {
		defer close(nextStage)
		for {
			select {
			case data, ok := <-prevStage:
				if !ok {
					slog.Warn("Канал предыдущей стадии закрыт", "stage", "NegativeFilter")
					return
				}

				slog.Debug("Получено число", "stage", "NegativeFilter", "value", data)

				if data >= 0 {
					select {
					case nextStage <- data:
						slog.Info("Число прошло фильтрацию", "stage", "NegativeFilter", "value", data)
					case <-done:
						slog.Warn("Получен сигнал завершения во время отправки", "stage", "NegativeFilter")
						return
					}
				} else {
					slog.Debug("Число не прошло фильтрацию", "stage", "NegativeFilter", "value", data)
				}
			case <-done:
				slog.Error("Стадия завершена по сигналу done", "stage", "NegativeFilter")
				return
			}
		}
	}()
	return nextStage
}

// Стадия фильтрации чисел не кратных трем
func NotDividedFilterStage(prevStage <-chan int, done <-chan struct{}) <-chan int {
	nextStage := make(chan int)

	go func() {
		defer close(nextStage)
		for {
			select {
			case data, ok := <-prevStage:
				if !ok {
					slog.Warn("Канал предыдущей стадии закрыт", "stage", "NotDividedFilter")
					return
				}

				slog.Debug("Получено число", "stage", "NotDividedFilter", "value", data)

				if data != 0 && data%3 == 0 {
					select {
					case nextStage <- data:
						slog.Info("Число прошло фильтрацию", "stage", "NotDividedFilter", "value", data)
					case <-done:
						slog.Warn("Получен сигнал завершения во время отправки", "stage", "NotDividedFilter")
						return
					}
				} else {
					slog.Debug("Число не прошло фильтрацию", "stage", "NotDividedFilter", "value", data)
				}
			case <-done:
				slog.Error("Стадия завершена по сигналу done", "stage", "NotDividedFilter")
				return
			}
		}
	}()
	return nextStage
}

// Стадия буферизации кольцевого буфера
func BufferStage(prevStage <-chan int, done <-chan struct{}) <-chan int {
	nextStage := make(chan int)
	buffer := NewRingBuffer(bufferSize)

	go func() {
		defer close(nextStage)
		for {
			select {
			case <-time.After(bufferDrainInterval):
				slog.Debug("Получение всех элементов в буфере", "stage", "Buffer")
				bufferData := buffer.Get()
				if len(bufferData) > 0 {
					slog.Info("Буфер сбрасывает данные", "stage", "Buffer", "value", bufferData)
					for _, data := range bufferData {
						select {
						case nextStage <- data:
						case <-done:
							slog.Warn("Получен сигнал завершения во время отправки", "stage", "Buffer")
							return
						}
					}
				} else {
					slog.Debug("Буфер пуст", "stage", "Buffer")
				}
			case <-done:
				slog.Error("Стадия завершена по сигналу done", "stage", "Buffer")
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case data, ok := <-prevStage:
				if !ok {
					slog.Warn("Канал предыдущей стадии закрыт", "stage", "Buffer")
					return
				}

				buffer.Push(data)
				slog.Debug("Добавлено в буфер", "stage", "Buffer", "value", data)
			case <-done:
				slog.Error("Стадия завершена по сигналу done", "stage", "Buffer")
				return
			}
		}
	}()
	return nextStage
}

// Обработка целых чисел в пайплайне
type StageInt func(prevStage <-chan int, done <-chan struct{}) <-chan int

// Структура пайплайна обработки целых чисел
type PipeLineInt struct {
	stages []StageInt
	done   <-chan struct{}
}

// Конструктор для пайплайна
func NewPipeLineInt(done <-chan struct{}, stages ...StageInt) *PipeLineInt {
	return &PipeLineInt{done: done, stages: stages}
}

// Запуск пайплайна
func (p *PipeLineInt) Run(source <-chan int) <-chan int {
	var s <-chan int = source
	for index := range p.stages {
		slog.Debug("Запуск стадии", "number", index+1)
		s = p.runStageInt(p.stages[index], s)
	}
	return s
}

// Запуск отдельной стадии пайплайна
func (p *PipeLineInt) runStageInt(stage StageInt, sourceChan <-chan int) <-chan int {
	return stage(sourceChan, p.done)
}
