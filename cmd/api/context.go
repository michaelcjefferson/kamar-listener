package main

import (
	"context"
	"net/http"

	"github.com/mjefferson-whs/listener/internal/data"
)

// Contexts allow storing data in key/value pairs for the lifetime of a request. Creating an app-specific context key type, as below, prevents clashes with other libraries that may be storing data in context key-values - if, for example, another library also uses the "user" key as below, it will not cause any problems, because it is of a different type (though both have a base type of string)
type contextKey string

const userContextKey = contextKey("user")

// Add the provdied user struct to the request's context, using "user" as the key (with the type of userContextKey)
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// Seeing as contextGetUser will only be called when we firmly expect a user to exist already in the request's context, it is ok to throw a panic when one does not in fact exist, as it is a very unexpected situation.
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}
