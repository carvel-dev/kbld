package tarball

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
)

type tarFile struct {
	path string
}

type tarFileChunk struct {
	file      tarFile
	chunkPath string
}

type tarFileChunkReadCloser struct {
	io.Reader
	io.Closer
}

func (f tarFile) Chunk(path string) tarFileChunk {
	return tarFileChunk{f, path}
}

func (f tarFileChunk) Open() (io.ReadCloser, error) {
	return f.file.openChunk(f.chunkPath)
}

func (f tarFile) openChunk(path string) (io.ReadCloser, error) {
	file, err := os.Open(f.path)
	if err != nil {
		return nil, err
	}
	tf := tar.NewReader(file)
	for {
		hdr, err := tf.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Name == path {
			return tarFileChunkReadCloser{Reader: tf, Closer: file}, nil
		}
	}
	return nil, fmt.Errorf("file %s not found in tar", path)
}
