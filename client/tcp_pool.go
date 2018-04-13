package client

import (
	"net"
	"fmt"
	log "github.com/Sirupsen/logrus"
	tcpPool "pool"
)
func buildTcpConn(conf map[string]interface{}) (interface{}, error) {
	return net.Dial("tcp", "127.0.0.1:80")
}

func closeTcpConn(conn interface{}, conf map[string]interface{}) error {
	return conn.(net.Conn).Close()
}

func main() {
	timeoutMs := 50
	poolSize := 100
	conf := map[string]interface{}{}
	pool, err := tcpPool.NewConnectionPool(buildTcpConn, closeTcpConn, timeoutMs, poolSize, conf)
	if err != nil {
		log.Errorf("init error : %v ",err)
	}
	conn, err :=pool.Get()
	if err != nil {
		log.Error("Get conn error : %v ",err)
	}
	fmt.Fprintf(conn.(net.Conn), "GET / HTTP/1.0\r\n\r\n")
	pool.Put(conn)
}