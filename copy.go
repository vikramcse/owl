// https://github.com/moby/moby/blob/master/daemon/graphdriver/copy/copy.go
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const buffer32K = 32 * 1024

var buffer32KPool = newBufferPoolWithSize(buffer32K)

type bufferPool struct {
	pool sync.Pool
}

func (bp *bufferPool) Get() *[]byte {
	return bp.pool.Get().(*[]byte)
}

func (bp *bufferPool) Put(b *[]byte) {
	bp.pool.Put(b)
}

func newBufferPoolWithSize(size int) *bufferPool {
	return &bufferPool{
		pool: sync.Pool{
			New: func() interface{} { s := make([]byte, size); return &s },
		},
	}
}

func poolCopy(dst io.Writer, src io.Reader) (written int64, err error) {
	buf := buffer32KPool.Get()
	written, err = io.CopyBuffer(dst, src, *buf)
	buffer32KPool.Put(buf)
	return
}

func fileCopy(srcf io.Reader, dest string, fileinfo os.FileInfo) error {
	dstPath := filepath.Join(dest, fileinfo.Name())
	dstf, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE, fileinfo.Mode())
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("owl: unable to create file (%v)", err)
	}
	defer dstf.Close()

	n, err := poolCopy(dstf, srcf)
	if err != nil {
		return fmt.Errorf("owl: unable to copy file (%v)", err)
	}
	fmt.Printf("%d bytes copied\n", n)

	// flush in-memory copy
	err = dstf.Sync()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
