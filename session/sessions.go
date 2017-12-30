package session

import (
	"sync"
	"crypto/rand"
	"net/http"
	"net/url"
	"io"
	"fmt"
	"encoding/base64"
	"time"
)

var provides = make(map[string]Provider)
var globalSessions *Manager
type Manager struct {
	cookieName  string     //private cookiename
	lock        sync.Mutex // protects session
	provider    Provider
	maxlifetime int64
}

//controls session features
type Provider interface {
	SessionInit(sid string) (Session, error)
	SessionRead(sid string) (Session, error)
	SessionDestroy(sid string) error
	SessionGC(maxLifeTime int64)
}

type Session interface {
	Set(key, value interface{}) error //set session value
	Get(key interface{}) interface{}  //get session value
	Delete(key interface{}) error     //delete session value
	SessionID() string                //back current sessionID
}

func NewManager(provideName, cookieName string, maxlifetime int64) (*Manager, error) {
	provider, ok := provides[provideName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", provideName)
	}
	return &Manager{provider: provider, cookieName: cookieName, maxlifetime: maxlifetime}, nil
}





// Register makes a session provider available by the provided name.
// If a Register is called twice with the same name or if the driver is nil,
// it panics.
func Register(name string, provider Provider) {
	if provider == nil {
		panic("session: Register provider is nil")
	}
	if _, dup := provides[name]; dup {
		panic("session: Register called twice for provider " + name)
	}
	provides[name] = provider
}


func (manager *Manager) sessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}


func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request) (session Session) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	cookie, err := r.Cookie(manager.cookieName)
	if err != nil || cookie.Value == "" {
		sid := manager.sessionId()
		session, _ = manager.provider.SessionInit(sid)
		cookie := http.Cookie{Name: manager.cookieName, Value: url.QueryEscape(sid), Path: "/", HttpOnly: true, MaxAge: int(manager.maxlifetime)}
		http.SetCookie(w, &cookie)
	} else {
		sid, _ := url.QueryUnescape(cookie.Value)
		session, _ = manager.provider.SessionRead(sid)
	}
	return
}

//
//func login(w http.ResponseWriter, r *http.Request) {
//	sess := globalSessions.SessionStart(w, r)
//	r.ParseForm()
//	if r.Method == "GET" {
//		t, _ := template.ParseFiles("login.gtpl")
//		w.Header().Set("Content-Type", "text/html")
//		t.Execute(w, sess.Get("username"))
//	} else {
//		sess.Set("username", r.Form["username"]) //is a POST
//		http.Redirect(w, r, "/", 302)
//	}
//}

func (manager *Manager) GC() {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.provider.SessionGC(manager.maxlifetime)
	time.AfterFunc(time.Duration(manager.maxlifetime), func() { manager.GC() })
}

//func init() {
//	globalSessions,_ = NewManager("memory","gosessionid",3600)
//}
