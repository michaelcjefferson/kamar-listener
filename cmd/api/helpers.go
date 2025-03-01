package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func (app *application) Render(c echo.Context, statusCode int, t templ.Component) error {
	c.Response().Writer.WriteHeader(statusCode)
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(c.Request().Context(), c.Response().Writer)
}

func (app *application) readIDParam(c echo.Context) (int, error) {
	param := c.Param("id")

	id, err := strconv.ParseInt(param, 10, 0)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return int(id), nil
}

type envelope map[string]interface{}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// js, err := json.Marshal(data)
	// MarshallIndent adds whitespace etc. to a JSON file so that it is nicely formatted and readable for terminal applications. This comes with a performance cost - best used either when performance is not an issue or when you know requests will regularly be made from the terminal. Ideally only use MarshalIndent on requests that are flagged as coming from terminal - look into this.
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// Appending a new line makes the response easier to read in terminal applications.
	js = append(js, '\n')
	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Use MaxBytesReader to limit the size of the request body to 50mb (the largest csv Mark has is 30mb, though this is from just one table (comments)). It closes the reader once it reaches maxBytes read and returns an error, attempting also to close the connection to the client
	maxBytes := 52_428_800
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)

	// Prevent unknown fields in the JSON body, returning an error if one is found rather than ignoring it
	// dec.DisallowUnknownFields()

	// Decode the request body into the target destination. This will be a pointer to an instantiated variable, in order to work with the Decode function
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// If error message starts with "json: unknown field", then get the field name by trimming the error message and return with a new error message. Currently there is a proposal to add a specific error (like MaxBytesError below) - check on this in the future
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// Lack of MaxBytesError was an open issue on Go's github which was resolved, hence the change
		// case err.Error() == "http: request body too large":
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	// Go's JSON decoder works with streams, so calling Decode() multiple times will look for additional elements to decode within the request body. As we only want one JSON object per request, we can call Decode() a second time, and if it is empty (and creates an io.EOF error), we have the expected single JSON object in the body, but in any other case return an error because there must be a second element in the request body.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// Take a function, and run it in a go routine with a deferred panic tidy-up.
func (app *application) background(fn func()) {
	// Use an application-level WaitGroup to keep track of all goroutines created by background(), by adding one to the WaitGroup counter before beginning the goroutine. When the app is shut down, use WaitGroup.Wait() to block until all background() goroutines have completed before terminating the app.
	app.wg.Add(1)

	go func() {
		// Decrement WaitGroup counter by 1 once fn() has returned.
		defer app.wg.Done()

		// Panics encountered in a goroutine will not be caught and tidied up by the recovery middleware. To remedy this, this deferred function will run after the fn() function below. By running recover(), it will catch any panics and log the error instead of terminating the application as would otherwise happen.
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()

		fn()
	}()
}
