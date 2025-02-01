package logger

import (
	"log"
	"os"
	"sync"
)

var (
	errorLogger *log.Logger
	infoLogger  *log.Logger
	debugLogger *log.Logger
	file        *os.File
	mu          sync.Mutex // Для потокобезпечного запису
)

// Init ініціалізує логер
func Init(logFileName string) {
	var err error
	file, err = os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Помилка при відкритті файлу логів: %v", err)
	}

	errorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	debugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Error логує повідомлення рівня ERROR
func Error(format string, v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	errorLogger.Printf(format, v...)
}

// Info логує повідомлення рівня INFO
func Info(format string, v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	infoLogger.Printf(format, v...)
}

// Debug логує повідомлення рівня DEBUG
func Debug(format string, v ...interface{}) {
	mu.Lock()
	defer mu.Unlock()
	debugLogger.Printf(format, v...)
}

// Close закриває файл логів
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		err := file.Close()
		if err != nil {
			return
		}
	}
}
