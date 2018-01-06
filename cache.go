package cache

import (
	"log"
	"github.com/go-crazy/cache/session"
)


var session *sessions.Sessions = sessions.New(sessions.Config{Cookie: "mycustomsessionid"})

func InitChche()  {
	session = sessions.New(sessions.Config{Cookie: "mycustomsessionid"})
}
func isInit() bool {
	if session != nil {
		log.Fatalf("Cache system is not init!")
		return false
	} 
	return true
}

func Cache(key string)   *sessions.Session {
	if isInit() {
		return session.Cache(key)
	}
	return nil
}

// UseDatabase adds a session database to the manager's provider,
// a session db doesn't have write access
func UseDatabase(db sessions.Database) {
	session.UseDatabase(db)
}
