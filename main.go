package main

import (
	"bufio"
	"fmt"
	"log"
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
	log.Println("Запуск конвейера")
	filterNegative := filterNegativeNumbers(input)
	filterDiv3 := filterDivisibleBy3(filterNegative)
	buffer := bufferedStage(filterDiv3, flushInterval)

	for value := range buffer {
		log.Printf("Конвейер передает значение: %d\n", value)
		output <- value
	}
	close(output)
	log.Println("Конвейер завершен")
}

// Фильтрация отрицательных чисел
func filterNegativeNumbers(input <-chan int) <-chan int {
	output := make(chan int)
	go func() {
		for value := range input {
			if value >= 0 {
				log.Printf("filterNegativeNumbers пропускает положительное значение: %d\n", value)
				output <- value
			} else {
				log.Printf("filterNegativeNumbers отфильтровал отрицательное значение: %d\n", value)
			}
		}
		close(output)
		log.Println("filterNegativeNumbers завершен")
	}()
	return output
}

// Фильтрация чисел, не кратных 3 (и исключение 0)
func filterDivisibleBy3(input <-chan int) <-chan int {
	output := make(chan int)
	go func() {
		for value := range input {
			if value != 0 && value%3 == 0 {
				log.Printf("filterDivisibleBy3 пропускает значение: %d\n", value)
				output <- value
			} else {
				log.Printf("filterDivisibleBy3 отфильтровал значение: %d\n", value)
			}
		}
		close(output)
		log.Println("filterDivisibleBy3 завершен")
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
					log.Println("Входной канал закрыт, опустошаем буфер и завершаем bufferedStage")
					flushBuffer(buffer, output)
					close(output)
					return
				}
				// Добавляем данные в буфер
				log.Printf("bufferedStage добавляет значение в буфер: %d\n", value)
				buffer = append(buffer, value)
				if len(buffer) >= bufferSize {
					log.Println("bufferedStage достиг размера буфера, опустошаем буфер")
					flushBuffer(buffer, output)
					buffer = buffer[:0]
				}
			case <-ticker.C:
				// Опустошаем буфер по таймеру
				if len(buffer) > 0 {
					log.Println("bufferedStage опустошает буфер по таймеру")
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
		log.Printf("flushBuffer передает значение: %d\n", value)
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

			log.Printf("inputSource получил значение: %d\n", value)
			output <- value
		}

		close(output)
		log.Println("inputSource завершен")
	}()
	return output
}

// Потребитель данных
func consumer(input <-chan int) {
	for value := range input {
		log.Printf("consumer получил значение: %d\n", value)
		fmt.Printf("Получены данные: %d\n", value)
	}
	log.Println("consumer завершен")
}

func main() {
	// Настройка логгера для вывода в консоль
	log.SetOutput(os.Stdout)
	log.Println("Запуск программы")

	// Создаем каналы
	input := inputSource()
	output := make(chan int)

	// Запускаем конвейер
	go pipeline(input, output)

	// Запускаем потребителя
	consumer(output)

	log.Println("Программа завершена")
}
