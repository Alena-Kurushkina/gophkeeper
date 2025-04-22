package gopherror

import "errors"

// ErrLoginAlreadyExists defines error in case of adding user with existed login
var ErrLoginAlreadyExists = errors.New("login is used by another user")

// ErrNoUserIDInToken defines error in case of empty user ID in JWT.
var ErrNoUserIDInToken = errors.New("no user ID in JWT")

// ErrTokenInvalid defines error in case of invalid JWT.
var ErrTokenInvalid = errors.New("token is not valid")

// ErrUnregisteredUser defines error in case of trying to login unregistered user
var ErrUnregisteredUser = errors.New("specified user hasn't found")

// ErrUnregisteredUser defines error in case of trying to login unregistered user
var ErrUnauthorized= errors.New("wrong login or password")

// ErrLoginAlreadyExists defines error in case of adding data with existed index key
var ErrAlreadyExists = errors.New("specified key already exists in storage")

// ErrLoginAlreadyExists defines error in case of getting empty data
var ErrNoData = errors.New("there is no data in storage")

// ErrAuthDecrypt defines error in case of message authentication failed while decrypt
var ErrDecryptAuth = errors.New("message authentication failed while decrypt")