package proxy

import (
  "net"
  "sync"
)

type SessionRegistry struct {
  // All session registry access must be locked.
  sync.RWMutex

  // Access by endpoint. The most common form of access.
  byEndpoint map[string]*Session

  // Access by username.
  byUsername map[string]*Session
}

func NewSessionRegistry() (this *SessionRegistry) {
	this = new(SessionRegistry)
	this.Clear()
	return
}

func (this *SessionRegistry) Register(session *Session) bool {
  this.Lock()
  defer this.Unlock()

  s := session.endpoint.String()

  _, ok := this.byEndpoint[s]
  if ok {
    return false
  }

  this.byEndpoint[s] = session

  return true
}

func (this *SessionRegistry) RegisterName(session *Session) bool {
  if session.username == nil {
    return false // Don't want nil usernames
  }

  this.Lock()
  defer this.Unlock()

  _, ok := this.byUsername[*session.username]
  if ok {
    return false
  }

  this.byUsername[*session.username] = session
  return true
}

func (this *SessionRegistry) Unregister(session *Session) {
  this.Lock()
  defer this.Unlock()

  delete(this.byEndpoint, session.endpoint.String())
  if session.username != nil {
    delete(this.byUsername, *session.username)
  }
}

func (this *SessionRegistry) GetByEndpoint(endpoint *net.UDPAddr) (session *Session) {
  this.RLock()
  defer this.RUnlock()
  return this.byEndpoint[endpoint.String()]
}

func (this *SessionRegistry) GetByUsername(username string) (session *Session) {
  this.RLock()
  defer this.RUnlock()
  return this.byUsername[username]
}

func (this *SessionRegistry) Clear() {
	this.byEndpoint = make(map[string]*Session)
	this.byUsername = make(map[string]*Session)
}

func (this *SessionRegistry) Len() (val int) {
  this.RLock()
  defer this.RUnlock()
	return len(this.byUsername)
}
