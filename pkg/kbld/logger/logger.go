// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

type Logger struct {
	writer     io.Writer
	writerLock *sync.Mutex
}

func NewLogger(writer io.Writer) Logger {
	return Logger{writer: writer, writerLock: &sync.Mutex{}}
}

func (l Logger) NewPrefixedWriter(prefix string) *PrefixWriter {
	return &PrefixWriter{prefix, l.writer, l.writerLock}
}

type PrefixWriter struct {
	prefix     string
	writer     io.Writer
	writerLock *sync.Mutex
}

func (w *PrefixWriter) Write(data []byte) (int, error) {
	newData := make([]byte, len(data))
	copy(newData, data)

	endsWithNl := bytes.HasSuffix(newData, []byte("\n"))
	if endsWithNl {
		newData = newData[0 : len(newData)-1]
	}
	newData = bytes.Replace(newData, []byte("\n"), []byte("\n"+w.prefix), -1)
	newData = append(newData, []byte("\n")...)
	newData = append([]byte(w.prefix), newData...)

	w.writerLock.Lock()
	defer w.writerLock.Unlock()

	// TODO does not deal with races of multitple writers
	_, err := w.writer.Write(newData)
	if err != nil {
		return 0, fmt.Errorf("write err: %s", err)
	}

	// return original data length
	return len(data), nil
}

func (w *PrefixWriter) WriteStr(str string, args ...interface{}) error {
	_, err := w.Write([]byte(fmt.Sprintf(str, args...)))
	return err
}
