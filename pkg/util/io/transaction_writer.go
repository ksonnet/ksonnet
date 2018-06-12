// Copyright 2018 The ksonnet authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package io

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// TransactionWriter implements a tranasactional, file-backed io.WriteCloser.
// Calls to Write() are captured in a temporary file, which will replace the
// target file path only upon Commit().
type TransactionWriter struct {
	io.WriteCloser
	tmpPath    string
	targetPath string
	fs         afero.Fs
}

// NewTransactionWriter returns a TransactionWriter implementing
// io.WritCloser. The caller is responsible for calling either
// Abort() or Commit().
// If Commit() is not called, the target file is not created/replaced.
func NewTransactionWriter(fs afero.Fs, path string) (*TransactionWriter, error) {
	if fs == nil {
		return nil, errors.Errorf("fs required")
	}
	tmp, err := afero.TempFile(fs, "", "kstemp-")
	if err != nil {
		return nil, err
	}

	return &TransactionWriter{
		WriteCloser: tmp,
		tmpPath:     tmp.Name(),
		targetPath:  path,
		fs:          fs,
	}, nil
}

// Abort aborts an in-progress transaction and cleans up any temporary files.
// This method is safe to call multiple times.
// Not currently thread-safe.
func (tw *TransactionWriter) Abort() error {
	if tw == nil {
		return nil
	}

	err := tw.Close()

	if tw.tmpPath == "" {
		return err
	}

	err = tw.fs.Remove(tw.tmpPath)
	tw.tmpPath = ""

	return err
}

// Commit commits an in-progress transaction and cleans up any temporary files.
// This method is safe to call multiple times.
// Not currently thread-safe.
func (tw *TransactionWriter) Commit() error {
	if tw == nil {
		return nil
	}

	if err := tw.Close(); err != nil {
		_ = tw.Abort()
		return err
	}

	if tw.tmpPath == "" {
		return nil
	}

	// Create target path structure
	targetDir := filepath.Dir(tw.targetPath)

	if err := tw.fs.MkdirAll(targetDir, os.FileMode(0755)); err != nil { // TODO pass permissions
		_ = tw.Abort()
		return errors.Wrapf(err, "failed creating target directory: %v", targetDir)
	}

	// Move temp file to correct location
	if err := tw.fs.Rename(tw.tmpPath, tw.targetPath); err != nil {
		_ = tw.Abort()
		return errors.Wrapf(err, "failed creating target directory: %v", targetDir)
	}

	return nil
}
