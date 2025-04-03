package gopherror

import "errors"

// ErrLoginAlreadyExists defines error in case of adding user with existed login
var ErrLoginAlreadyExists = errors.New("login is used by another user")