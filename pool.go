package pool

import (
	"sync"
	log "github.com/Sirupsen/logrus"
	"errors"
)

type buildFunction func(map[string]interface{}) (interface{}, error)
type closeFunction func(interface{}, map[string]interface{}) error
type Pool struct {
	conns       chan interface{}
	build       buildFunction
	close       closeFunction
	timeoutMs   int
	poolSize    int
	mu          sync.Mutex
	extra       map[string]interface{}
} 

func NewConnectionPool(build buildFunction, close closeFunction, timeoutMs int, poolSize int, extra map[string]interface{}) (pool *Pool,err error) {
	pool = &Pool{
		conns:make(chan interface{},poolSize),
		build:build,
		close:close,
		timeoutMs:timeoutMs,
		poolSize:poolSize,
		extra:extra,
	}
	for i :=0; i<poolSize; i++ {
		conn, err := pool.build(pool.extra)
		if err != nil {
			pool.Release()
			return nil,errors.New("[connection pool] Init connection pool error")
		}
		pool.conns <- conn
	}
	return
}

func (pool *Pool) Get() (conn interface{}, err error) {
	err = nil
	ok := true
	pool.mu.Lock()
	if len(pool.conns) != 0 {
		conn, ok = <-pool.conns
		pool.mu.Unlock()
		if ok && conn !=nil {
			return
		} else {
			return nil, errors.New("[connection pool] Get error,The pool is closed")
		}
	} else {
		return nil, errors.New("[connection pool] Get error,The pool is empty")
	}
}

func (pool *Pool) Put(conn interface{}) (err error) {
	err = nil
	if conn == nil && pool.build != nil {
		conn, err = pool.build(pool.extra)
		if err != nil {
			err = errors.New("[connection pool] Put error,can't build a new conn")
			return
		}
	}
	defer func() {
		if recover() != nil {
			err = errors.New("[connection pool] Put error,The pool is closed")
		}
	}()
	pool.conns <- conn
	return
}

func (pool *Pool) Close(conn interface{}) (err error) {
	err = pool.close(conn, pool.extra)
	conn = nil
	return
}

func (pool *Pool) Release() {
	log.Infof("[connection pool] Release the connection pool")
	defer func() {
		if recover() != nil {
			log.Error("[connection pool] Release error,The pool has been closed")
		}
	}()
	pool.build = nil
	close(pool.conns)
	for conn := range pool.conns {
		pool.mu.Lock()
		pool.close(conn, pool.extra)
		conn = nil
		pool.mu.Unlock()
	}
}