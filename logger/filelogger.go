package logger

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"sync"
)

type FileTransactionLogger struct {
	events       chan<- Event
	errors       <-chan error
	lastSequence uint64
	file         *os.File
	wg           *sync.WaitGroup
}

const formatString = "%d\t%d\t%s\t%s\n"

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
		wg:   &sync.WaitGroup{},
	}
	return &ftl, nil
}

func (l *FileTransactionLogger) Run() {
	//TODO: make sure this is hiding deadlocks
	events := make(chan Event, 16)
	l.events = events

	errors := make(chan error, 1)
	l.errors = errors

	go func() {
		for e := range events {
			l.lastSequence++
			_, err := fmt.Fprintf(
				l.file,
				formatString,
				l.lastSequence, e.EventType, e.Key, e.Value)
			if err != nil {
				errors <- err
			}
			l.wg.Done()
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

			_, err := fmt.Sscanf(
				line, formatString,
				&e.Sequence, &e.EventType, &e.Key, &e.Value)
			if err != nil {
				outError <- err
				return
			}
			if l.lastSequence >= e.Sequence {
				outError <- fmt.Errorf("transaction numbers out of sequence")
				return
			}

			uv, err := url.QueryUnescape(e.Value)
			if err != nil {
				outError <- fmt.Errorf("value decoding failurefor Value '%s':\t%w", e.Value, err)
				return
			}
			e.Value = uv
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
func (l *FileTransactionLogger) Wait() {
	l.wg.Wait()
}

func (l *FileTransactionLogger) Close() error {
	l.Wait()

	if l.events != nil {
		close(l.events)
	}

	return l.file.Close()
}

func (l *FileTransactionLogger) WritePut(key, value string) {
	l.wg.Add(1)
	l.events <- Event{
		EventType: EventPut,
		Key:       key,
		Value:     value,
	}
}

func (l *FileTransactionLogger) WriteDelete(key string) {
	l.wg.Add(1)
	l.events <- Event{
		EventType: EventDelete,
		Key:       key,
	}
}

func (l *FileTransactionLogger) Err() <-chan error {
	return l.errors
}
func (l *FileTransactionLogger) LastSequence() uint64 {
	return l.lastSequence
}
