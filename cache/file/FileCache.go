package file

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
)

type Files interface {
	Read(file *File) (bytes []byte, err error)
	Write(file *File, bytes []byte) (err error)
}

type File struct {
	File  *os.File
	Rw    sync.RWMutex //读写锁
	Cond  *sync.Cond   //条件锁
	IsEof uint32
}

type Read struct {
	Offset  uint64 //偏移量
	ReadLen uint64 //读取长度
	Rsn     uint64 //读取计数

}

type Write struct {
	Offset   uint64 //偏移量
	WriteLen uint64 //写入长度
	Wsn      uint64 //写入计数

}

func NewRead(readLen uint64) (*Read, error) {
	if readLen == 0 {
		return nil, errors.New("读取长度不能为空")
	}
	return &Read{Offset: 0, ReadLen: readLen}, nil
}

func NewWrite(writeLen uint64) (*Write, error) {
	if writeLen == 0 {
		return nil, errors.New("写入长度不能为空")
	}
	return &Write{Offset: 0, WriteLen: writeLen}, nil
}

func NewFile(F *os.File) (*File, error) {
	if F == nil {
		return nil, errors.New("F 不能为空")
	}
	f := &File{File: F}

	f.Cond = sync.NewCond(&sync.Mutex{})
	return f, nil
}

func (r *Read) Read(file *File) (bytes []byte, err error) {
	var (
		offset    uint64
		newOffset uint64
	)

	bytes = make([]byte, r.ReadLen)

	for {
		//添加读写锁
		file.Rw.Lock()
		fmt.Println("read 获取到读写锁：", os.Getpid())

		//原子方法加载offset的值
		offset = atomic.LoadUint64(&r.Offset)

		//计算新值
		newOffset = offset + r.ReadLen

		// 原子方法更新offset的值(CAS: 如果被操作的值&r.Offset等于旧值offset, 就将他的值更新为newOffset)
		if atomic.CompareAndSwapUint64(&r.Offset, offset, newOffset) {

			//从offset开始读取bytes个字节
			if _, err = file.File.ReadAt(bytes, int64(offset)); err != nil {
				//如果文件没有读完
				if err != io.EOF {
					fmt.Println("read 读取错误：", err)
					return
				} else {
					//原子更新file.IsEof
					if atomic.CompareAndSwapUint32(&file.IsEof, 0, 1) {
						//等待写入文件信号
						fmt.Println("等待写入1*******************************************************", os.Getpid())
						file.Rw.Unlock()
						fmt.Println("read 释放读取锁", os.Getpid())
						file.Cond.L.Lock()
						file.Cond.Wait()
						file.Cond.L.Unlock()
						fmt.Println("等待写入2*******************************************************")
						continue

					}
					file.Rw.Unlock()
					fmt.Println("read 释放读取锁", os.Getpid())
					continue
				}
			} else {
				//原子方法加1
				atomic.AddUint64(&r.Rsn, 1)
				break
			}
		}

	}
	//释放读写锁
	defer func() {
		if errs := recover(); errs != nil {
			err = errors.New(fmt.Sprintf("read err :%s", errs))
		}
		file.Rw.Unlock()
		fmt.Println(" defer read 释放读写锁", os.Getpid())
	}()
	return

}

func (w *Write) Write(file *File, bytes []byte) (err error) {
	var (
		offset    uint64
		newOffset uint64
	)

	for {
		file.Rw.Lock()
		fmt.Println("write 获取到读取锁")
		//原子方法获取写的偏移量offset
		offset = atomic.LoadUint64(&w.Offset)
		//计算偏移量newOffset
		newOffset = offset + w.WriteLen
		//原子方法更新偏移量Offset
		if atomic.CompareAndSwapUint64(&w.Offset, offset, newOffset) {
			if _, err = file.File.WriteAt(bytes, int64(offset)); err != nil {
				return errors.New(fmt.Sprintf("写入文件错误：%s", err))
			} else {
				//原子方法更新写入计数
				atomic.AddUint64(&w.Wsn, 1)
				if atomic.CompareAndSwapUint32(&file.IsEof, 1, 0) {
					//发送写入信号
					fmt.Println("发送写入信号1************************************")
					file.Cond.Signal()
					fmt.Println("发送写入信号2****************************************")
				}

				break
			}
		} else {
			file.Rw.Unlock()
		}
	}
	//释放读写锁
	defer func() {
		if errs := recover(); errs != nil {
			err = errors.New(fmt.Sprintf("write err :%s", errs))
		}
		file.Rw.Unlock()
		fmt.Println("write 释放读写锁")
	}()
	return nil

}
