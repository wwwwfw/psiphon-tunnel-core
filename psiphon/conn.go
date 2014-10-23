/*
 * Copyright (c) 2014, Psiphon Inc.
 * All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package psiphon

import (
	"net"
	"sync"
)

// Dialer is a custom dialer compatible with http.Transport.Dial.
type Dialer func(string, string) (net.Conn, error)

// Conn is a net.Conn which supports sending a signal to a channel when
// it is closed. In Psiphon, this interface is implemented by tunnel
// connection types (DirectConn and MeekConn) and the close signal is
// used as one trigger for tearing down the tunnel.
type Conn interface {
	net.Conn

	// SetClosedSignal sets the channel which will be signaled
	// when the connection is closed. This function returns an error
	// if the connection is already closed (and would never send
	// the signal). SetClosedSignal and Close may be called by
	// concurrent goroutines.
	SetClosedSignal(closedSignal chan struct{}) (err error)
}

// Conns is a synchronized list of Conns that is used to coordinate
// interrupting a set of goroutines establishing connections, or
// close a set of open connections, etc.
type Conns struct {
	mutex sync.Mutex
	conns map[net.Conn]bool
}

func (conns *Conns) Add(conn net.Conn) {
	conns.mutex.Lock()
	defer conns.mutex.Unlock()
	if conns.conns == nil {
		conns.conns = make(map[net.Conn]bool)
	}
	conns.conns[conn] = true
}

func (conns *Conns) Remove(conn net.Conn) {
	conns.mutex.Lock()
	defer conns.mutex.Unlock()
	delete(conns.conns, conn)
}

func (conns *Conns) CloseAll() {
	conns.mutex.Lock()
	defer conns.mutex.Unlock()
	for conn, _ := range conns.conns {
		conn.Close()
	}
	conns.conns = make(map[net.Conn]bool)
}