package logger

import (
	"bufio"
	"fmt"
	"os"
)

type FileTransactionLogger struct {
	events       chan<- Event
	errors       <-chan error
	lastSequence uint64
	file         *os.File
}

func NewFileTransactionLogger(filename string) (*FileTransactionLogger, error) {
	//os.O_RDWR Opens the file in r/w mode
	//os.O_APPEND All writes will append, not overwrite
	//os.O_CREATE If the file doesn't exit, create it
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open translation log file %s: %w", filename, err)
	}
	ftl := FileTransactionLogger{
		file: file,
	}
	return &ftl, nil
}

func (l *FileTransactionLogger) Run() {
	//TODO: make sure this is hiding deadlocks
	events := make(chan Event, 16)
	errors := make(chan error, 1)
	l.events = events
	l.errors = errors
	go func() {
		for e := range events {
			l.lastSequence++
			_, err := fmt.Fprintf(
				l.file,
				"%d|\t%d|\t%s|\t%s\n",
				l.lastSequence, e.EventType, e.Key, e.Value)
			if err != nil {
				errors <- err
				return
			}
		}
	}()
}
func (l *FileTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	scanner := bufio.NewScanner(l.file)
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		var e Event

		defer close(outEvent)
		defer close(outError)

		for scanner.Scan() {
			line := scanner.Text()

			_, err := fmt.Sscanf(line, "%d\t%d\t%s\t%s", &e.Sequence, &e.EventType, &e.Key, &e.Value)
			if err != nil {
				outError <- fmt.Errorf("inpurt parse error: %w", err)
				return
			}
			if l.lastSequence >= e.Sequence {
				outError <- fmt.Errorf("transaction numbers out of sequence")
				return
			}

			l.lastSequence = e.Sequence

			outEvent <- e
		}

		if err := scanner.Err(); err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
			return
		}
	}()

	return outEvent, outError
}
func (l *FileTransactionLogger) WritePut(key, value string) {
	l.events <- Event{
		EventType: EventPut,
		Key:       key,
		Value:     value,
	}
}
func (l *FileTransactionLogger) WriteDelete(key string) {
	l.events <- Event{
		EventType: EventDelete,
		Key:       key,
	}
}
func (l *FileTransactionLogger) Err() <-chan error {
	return l.errors
}
