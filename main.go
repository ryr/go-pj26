package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	bufferSize    = 5               // Размер буфера
	flushInterval = 3 * time.Second // Интервал опустошения буфера
)

// Конвейер
func pipeline(input <-chan int, output chan<- int) {
	filterNegative := filterNegativeNumbers(input)
	filterDiv3 := filterDivisibleBy3(filterNegative)
	buffer := bufferedStage(filterDiv3, flushInterval)

	for value := range buffer {
		output <- value
	}
	close(output)
}

// Фильтрация отрицательных чисел
func filterNegativeNumbers(input <-chan int) <-chan int {
	output := make(chan int)
	go func() {
		for value := range input {
			if value >= 0 {
				output <- value
			}
		}
		close(output)
	}()
	return output
}

// Фильтрация чисел, не кратных 3 (и исключение 0)
func filterDivisibleBy3(input <-chan int) <-chan int {
	output := make(chan int)
	go func() {
		for value := range input {
			if value != 0 && value%3 == 0 {
				output <- value
			}
		}
		close(output)
	}()
	return output
}

// Буферизация данных
func bufferedStage(input <-chan int, flushInterval time.Duration) <-chan int {
	output := make(chan int)
	buffer := make([]int, 0, bufferSize)

	go func() {
		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()

		for {
			select {
			case value, ok := <-input:
				if !ok {
					// Если входной канал закрыт, опустошаем буфер
					flushBuffer(buffer, output)
					close(output)
					return
				}
				// Добавляем данные в буфер
				buffer = append(buffer, value)
				if len(buffer) >= bufferSize {
					flushBuffer(buffer, output)
					buffer = buffer[:0]
				}
			case <-ticker.C:
				// Опустошаем буфер по таймеру
				if len(buffer) > 0 {
					flushBuffer(buffer, output)
					buffer = buffer[:0]
				}
			}
		}
	}()

	return output
}

// Функция для опустошения буфера
func flushBuffer(buffer []int, output chan<- int) {
	for _, value := range buffer {
		output <- value
	}
}

// Источник данных - чтение из консоли
func inputSource() <-chan int {
	output := make(chan int)
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Введите целые числа (для завершения ввода введите '(e)xit'):")

		for scanner.Scan() {
			text := scanner.Text()
			if text == "exit" || text == "e" {
				break
			}

			// Пробуем преобразовать ввод в целое число
			value, err := strconv.Atoi(text)
			if err != nil {
				fmt.Println("Введите целое число")
				continue
			}

			output <- value
		}

		close(output)
	}()
	return output
}

// Потребитель данных
func consumer(input <-chan int) {
	for value := range input {
		fmt.Printf("Получены данные: %d\n", value)
	}
}

func main() {
	// Создаем каналы
	input := inputSource()
	output := make(chan int)

	// Запускаем конвейер
	go pipeline(input, output)

	// Запускаем потребителя
	consumer(output)
}
