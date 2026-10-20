package realtime

import (
	"fmt"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type sessionUsers interface {
	SignedInUser(r *http.Request) (domain.User, bool, error)
}

type lazyUser struct {
	users    sessionUsers
	request  *http.Request
	looked   bool
	user     domain.User
	signedIn bool
}

func newLazyUser(users sessionUsers, request *http.Request) *lazyUser {
	var nobody domain.User

	return &lazyUser{users: users, request: request, looked: false, user: nobody, signedIn: false}
}

func (l *lazyUser) get() (domain.User, bool, error) {
	if !l.looked {
		user, signedIn, err := l.users.SignedInUser(l.request)
		if err != nil {
			return user, false, fmt.Errorf("look up the signed-in user: %w", err)
		}

		l.user, l.signedIn, l.looked = user, signedIn, true
	}

	return l.user, l.signedIn, nil
}
