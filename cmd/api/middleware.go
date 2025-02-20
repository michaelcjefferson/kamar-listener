package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mjefferson-whs/listener/internal/data"
	"github.com/mjefferson-whs/listener/internal/validator"
)

func (app *application) recoverPanicMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Deferring a function ensures it will execute as the stack is unwound in the case of a panic. It won't be used elsewhere, so an anonymous function works well.
		defer func() {
			// recover() is a built-in function that checks whether or not there has been a panic.
			if err := recover(); err != nil {
				// Set header which tells the server to close the connection after this has been sent.
				c.Response().Header().Set("Connection", "close")
				app.logger.PrintError(err.(error), map[string]interface{}{
					"OHNO": "couldn't recover from this error",
				})
				c.Error(fmt.Errorf("%v", err))
			}
		}()

		return next(c)
	}
}

// func (app *application) recoverPanic(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Deferring a function ensures it will execute as the stack is unwound in the case of a panic. It won't be used elsewhere, so an anonymous function works well.
// 		defer func() {
// 			// recover() is a built-in function that checks whether or not there has been a panic.
// 			if err := recover(); err != nil {
// 				// Set header which tells the server to close the connection after this has been sent.
// 				w.Header().Set("Connection", "close")

// 				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
// 			}
// 		}()

// 		next.ServeHTTP(w, r)
// 	})
// }

// // This will not work on a system running multiple servers. Two options - if using a load balancer like Nginx, use its built-in rate-limiting functionality. Alternatively, run a speedy database like Redis on a separate server that all other servers communicate with, and hold the clients map there.
// func (app *application) rateLimit(next http.Handler) http.Handler {
// 	type client struct {
// 		limiter  *rate.Limiter
// 		lastSeen time.Time
// 	}

// 	// Create a map to hold clients' individual limiters and lastSeen time, and a mutex to prevent race conditions (as multiple goroutines may be accessing the map at once)
// 	var (
// 		mu      sync.Mutex
// 		clients = make(map[string]*client)
// 	)

// 	// Background goroutine which removes clients from clients map that haven't been seen for at least 3 minutes, once every minute
// 	go func() {
// 		for {
// 			time.Sleep(time.Minute)

// 			// Delay any rate checks etc. while clean up is happening
// 			mu.Lock()

// 			for ip, client := range clients {
// 				if time.Since(client.lastSeen) > 3*time.Minute {
// 					delete(clients, ip)
// 				}
// 			}

// 			mu.Unlock()
// 		}
// 	}()

// 	// Return a closure which creates a perpetuated clients map etc. when rateLimit is called (when the server is initialised), and allows the returned handler function to run each time a request is made - this function manipulates the clients map
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if app.config.limiter.enabled {
// 			app.logger.PrintInfo(fmt.Sprintf("%s request received from %s", r.Method, r.RemoteAddr), nil)

// 			ip, _, err := net.SplitHostPort(r.RemoteAddr)
// 			if err != nil {
// 				app.serverErrorResponse(w, r, err)
// 				return
// 			}

// 			mu.Lock()

// 			if _, found := clients[ip]; !found {
// 				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
// 			}

// 			clients[ip].lastSeen = time.Now()

// 			if !clients[ip].limiter.Allow() {
// 				mu.Unlock()
// 				app.rateLimitExceededResponse(w, r)
// 				return
// 			}

// 			// Don't defer unlocking of this mutex, as it will cause it to wait until handlers wrapped by this function eg. GET /v1/movies have returned before unlocking
// 			mu.Unlock()
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }

// Authenticate requests received from KAMAR itself, using the required Basic authentication
func (app *application) authenticateKAMAR(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.logger.PrintInfo("failed at authHeader", nil)
			app.noCredentialsResponse(w, r)
			return
		}

		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Basic" {
			app.logger.PrintInfo("failed at headerParts", nil)
			app.authFailedResponse(w, r)
			return
		}

		decodedAuth, err := base64.StdEncoding.DecodeString(headerParts[1])
		if err != nil {
			app.logger.PrintInfo("failed at decodedAuth", nil)
			app.authFailedResponse(w, r)
			return
		}

		authCredentials := strings.Split(string(decodedAuth), ":")
		if len(authCredentials) != 2 || authCredentials[0] != app.config.credentials.username || authCredentials[1] != app.config.credentials.password {
			logInfo := make(map[string]interface{})
			logInfo["creds"] = authCredentials
			logInfo["creds_length"] = len(authCredentials)
			logInfo["app_user"] = app.config.credentials.username
			logInfo["app_pass"] = app.config.credentials.password
			logInfo["req_user"] = authCredentials[0]
			logInfo["req_pass"] = authCredentials[1]
			app.logger.PrintInfo("failed at authCredentials", logInfo)
			app.authFailedResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// If a valid auth token is provided, set "user" value in request context to a struct containing the corresponding user's data. If an invalid token is provided, send an error.
func (app *application) authenticateUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the below logic when Authorization headers are required for auth, eg. from mobile apps. For now, as it is a browser-based app, http cookies are a better choice
		// // The Vary: Authorization header indicates to any (browser?) caches that the response may Vary based on Authorization provided in the request, and in doing so perhaps prevents cached data from no-longer-valid authorization from being loaded.
		// w.Header().Add("Vary", "Authorization")

		// // Returns "" if no Authorization header found in request
		// authorizationHeader := r.Header.Get("Authorization")

		// // In the above case, set request context with an AnonymousUser user value
		// if authorizationHeader == "" {
		// 	r = app.contextSetUser(r, data.AnonymousUser)
		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		// // The token is expected to be provided in the Authorization header in the format "Bearer <token>", so attempt to split the header to isolate the token, and if the result is unexpected, return an error.
		// headerParts := strings.Split(authorizationHeader, " ")
		// if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		// 	app.invalidAuthenticationTokenResponse(w, r)
		// 	return
		// }

		// token := headerParts[1]

		// TODO: Add a check for r.Context().Value("isAuthenticated").(bool) to prevent extra look-ups if user is already authenticated, and set the isAuthenticated value below once user has been found. Ensure that isAuthenticated doesn't lead to leaky security, where this can be parsed as true even if user has no or an expired token.

		// Get the http-only cookie containing the token from the request, and convert to a string
		cookie, err := r.Cookie("listener_admin_auth_token")

		// If the cookie can't be found, the user is not authenticated and should be set as an anonymous user
		if err == http.ErrNoCookie {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		token := cookie.Value

		v := validator.New()

		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Retrieve user data from user table based on the token provided.
		user, tokenExpiry, err := app.models.Users.GetForToken(token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// If token has only a short time before expiry, create a new token for that user
		expiryTime, err := time.Parse(time.RFC3339, tokenExpiry)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		expiryTimeFrame := time.Now().Add(app.config.tokens.refresh)

		// Check if the token expiry is within the timeframe, and if so, generate a new token and return it
		if expiryTime.Before(expiryTimeFrame) {
			app.logger.PrintInfo("token near expiry - creating new token and sending to user", map[string]interface{}{
				"user id":           user.ID,
				"expiry time":       tokenExpiry,
				"expiry time frame": expiryTimeFrame,
			})
			app.createAndSetAdminTokenCookie(w, user.ID, app.config.tokens.expiry)
		}

		// Attach user data to context
		r = app.contextSetUser(r, user)

		// Call next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// TODO: Add activation and requireActivatedUser for new user registrations after initial set-up - admins can go to an add user page, and enter an email address to send an activation code to. This creates an activation token in the database, and provides it as part of a link for the admin to copy and paste into an email to the new user. The new user can follow that link to be brought to an activation page, where they create a username and password, and a new account is created. Activation tokens valid for 24 (?) hours

// Runs after authenticate, only needed on protected routes - checks the context for the value of the user set by authenticate, and at this point only ensures that one exists, as it means that someone is logged in and can access protected routes
func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// // TODO: Add to config for app, including instructions to find IP address of KAMAR instance
// func (app *application) processCORS(next http.Handler) http.Handler {
// 	c := cors.New(cors.Options{
// 		AllowedOrigins: []string{"https://localhost", "https://0.0.0.0"},
// 		// AllowedOrigins:   []string{"https://localhost", "https://10.100"},
// 		AllowCredentials: true,
// 		AllowedHeaders:   []string{"Origin", "Authorization", "Content-Type"},
// 		AllowedMethods:   []string{"GET", "POST"},
// 		// AllowedMethods:   []string{"POST"},
// 		Debug: true,
// 	})

// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		c.Handler(next).ServeHTTP(w, r)
// 	})
// }
