# Connector

Connector is an opinionated HTTP server which provides a Handle method for defining routes.

Routes must be defined according to the schema: HTTP method, resource, middleware(s) and/or handler.

Middlewares must implement the *Do* interface:

```go
	Do(context.Context, http.ResponseWriter) context.Context
```

The context passed to the middlewares, which can be read and expanded from one to the next, includes:

- Body
- Cookies
- Identifier
- HTTP Method
- Query
- Resource

For examples of middlewares:

* [**Logger**](https://github.com/attheapplab/logger-go) for logging incoming HTTP requests.

## Example

```go
package main

import (
	"github.com/attheapplab/connector-go"
	"net/http"
)

type middleware struct {}

func New() *middleware {
	return &middleware{}
}

func (m *middleware) Do(ctx context.Context, w http.ResponseWriter) context.Context {
	// Do anything you want in your awesome middleware...
	// ...finally return the context.
	return ctx
}


func main() {
	// Create a new instance of Connector handler.
	handler := connector.New()

	// Create your middleware instances.
	mw1 := middleware.New()
	mw2 := middleware.New()
	mw3 := middleware.New()
	
	// Define the route with HTTP method, resource, middleware(s) and/or handler.
	handler.Handle(http.MethodGet, "first", mw1)
	handler.Handle(http.MethodGet, "second", mw2)
	handler.Handle(http.MethodGet, "third", mw3)
	handler.Handle(http.MethodGet, "fourth", mw1, mw2, mw3)
	handler.Handle(http.MethodPost, "fifth", mw1, mw2, mw3)
	handler.Handle(http.MethodDelete, "sixth", mw1, mw2, mw3)
	
	// Start the server.
	handler.ListenAndServe()
}

```

## Contributing
For major changes, please open an issue first to discuss what you would like to change.

## License
[MIT](https://choosealicense.com/licenses/mit/)