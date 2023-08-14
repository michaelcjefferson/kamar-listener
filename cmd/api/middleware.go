package main

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Deferring a function ensures it will execute as the stack is unwound in the case of a panic. It won't be used elsewhere, so an anonymous function works well.
		defer func() {
			// recover() is a built-in function that checks whether or not there has been a panic.
			if err := recover(); err != nil {
				// Set header which tells the server to close the connection after this has been sent.
				w.Header().Set("Connection", "close")

				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// This will not work on a system running multiple servers. Two options - if using a load balancer like Nginx, use its built-in rate-limiting functionality. Alternatively, run a speedy database like Redis on a separate server that all other servers communicate with, and hold the clients map there.
func (app *application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// Create a map to hold clients' individual limiters and lastSeen time, and a mutex to prevent race conditions (as multiple goroutines may be accessing the map at once)
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Background goroutine which removes clients from clients map that haven't been seen for at least 3 minutes, once every minute
	go func() {
		for {
			time.Sleep(time.Minute)

			// Delay any rate checks etc. while clean up is happening
			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	// Return a closure which creates a perpetuated clients map etc. when rateLimit is called (when the server is initialised), and allows the returned handler function to run each time a request is made - this function manipulates the clients map
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			// Don't defer unlocking of this mutex, as it will cause it to wait until handlers wrapped by this function eg. GET /v1/movies have returned before unlocking
			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.failedResponse(w, r)
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Basic" {
			app.failedResponse(w, r)
			return
		}

		decodedAuth, err := base64.StdEncoding.DecodeString(headerParts[1])
		if err != nil {
			app.failedResponse(w, r)
			return
		}

		authCredentials := strings.Split(string(decodedAuth), ":")
		if len(authCredentials) != 2 || authCredentials[0] != app.config.credentials.username || authCredentials[1] != app.config.credentials.password {
			app.failedResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
