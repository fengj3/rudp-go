package rudp

import (
	"net"
	"sync"
)

func NewListener() *RudpListener {
	listen := &RudpListener{
		newRudpConn: make(chan *RudpConn, 1024),
		newRudpErr:  make(chan error, 12),
		rudpConnMap: make(map[string]*RudpConn)}
	go listen.run()
	return listen
}

type RudpListener struct {
	lock sync.RWMutex

	newRudpConn chan *RudpConn
	newRudpErr  chan error
	rudpConnMap map[string]*RudpConn
}

//net listener interface
func (this *RudpListener) Accept() (*RudpConn, error) { return this.AcceptRudp() }
func (this *RudpListener) Close() {
	this.CloseAllRudp()
}
func (this *RudpListener) Addr() net.Addr { return (*pc).LocalAddr() }

func (this *RudpListener) CloseRudp(addr string) {
	this.lock.Lock()
	delete(this.rudpConnMap, addr)
	this.lock.Unlock()
}

func (this *RudpListener) CloseAllRudp() {
	this.lock.Lock()
	for _, rconn := range this.rudpConnMap {
		rconn.Close()
	}
	this.lock.Unlock()
}
func (this *RudpListener) AcceptRudp() (*RudpConn, error) {
	select {
	case c := <-this.newRudpConn:
		return c, nil
	case e := <-this.newRudpErr:
		return nil, e
	}
}
func (this *RudpListener) run() {
	data := make([]byte, MAX_PACKAGE)
	for {
		n, remoteAddr, err := (*pc).ReadFrom(data)
		if err != nil {
			this.CloseAllRudp()
			this.newRudpErr <- err
			return
		}
		this.lock.RLock()
		rudpConn, ok := this.rudpConnMap[remoteAddr.String()]
		this.lock.RUnlock()
		if !ok {
			rudpConn = NewConn(AddrToUDPAddr(&remoteAddr), NewRudp())
			this.lock.Lock()
			this.rudpConnMap[remoteAddr.String()] = rudpConn
			this.lock.Unlock()
			this.newRudpConn <- rudpConn
		}
		bts := make([]byte, n)
		copy(bts, data[:n])
		rudpConn.in <- bts
	}
}