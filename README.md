# Connector

Connector is an opinionated HTTP router with easy drop-in of middleware and handlers.

## Routes

Define your routes using HTTP verbs (Delete, Get, Patch, Post) and a variable number of middleware(s) and final handlers.

```golang
func main() {
    conn := connector.New()
   	conn.Delete("/", DeleteHandler)
    conn.Get("/", GetHandler)
    conn.Patch("/", PatchHandler)
    conn.Post("/", PostHandler)
}
```

When a route is matched, each handler will have access to the same instance of context.

### Context

Read and write the context which is passed along with every request. By default the context includes the following keys:

- **method** `string`
> A verb (like GET, PUT or POST) or a noun (like HEAD or OPTIONS), that describes the action to be performed.

- **body** `map[string]interface{}`
> Not all requests have a body: requests fetching resources, like GET, HEAD, DELETE, or OPTIONS, usually don't need one.

- **query** `map[string]interface{}`
> Query parameters, if any.

- **cookies** `map[string]string`
> Cookies, if any.

- **resource** `string`
> Resource part of the request path i.e. http://www.example.com/{resource}/{identifier}

- **identifier** `string`
> Identifier part of the request path i.e. http://www.example.com/{resource}/{identifier}

## Middleware

Important! Middleware and handlers must implement the *Do* interface:

```golang
	Do(context.Context, http.ResponseWriter) context.Context
```

An example of a simple handler implementing the *Do* interface could be:

```golang
type greeter struct {}

func (g *greeter) Do(ctx context.Context, w http.ResponseWriter) context.Context {
	// Do anything you need here...
	resource, _ := ctx.Value("resource").(string)
	fmt.Println("Hello", resource, "!")
	// ...finally return the context.
	return ctx
}
```

Examples of public middleware and handlers:

- [**Logger**](https://github.com/attheapplab/logger-go) to log incoming HTTP requests.
- [**Postregd**](https://github.com/attheapplab/postregd-go) to sign-up (insert) users in a PostgreSQL database.

## Hello World Example

```golang
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
	// Do anything you need here...
	resource, _ := ctx.Value("resource").(string)
	fmt.Println("Hello", resource, "!")
	// ...finally return the context.
	return ctx
}

func main() {
	// Create a new instance of Connector.
	conn := connector.New()

	// Create your middleware instances.
	mw1 := middleware.New()
	mw2 := middleware.New()
	mw3 := middleware.New()
	
	// Define the route with HTTP method, resource, middleware(s) and/or handler.
	conn.Get("first", mw1)
	conn.Get("second", mw2)
	conn.Get("third", mw3)
	conn.Get("fourth", mw1, mw2, mw3)
	conn.Post("fifth", mw1, mw2, mw3)
	conn.Delete("sixth", mw1, mw2, mw3)
	
	// Start the server.
	conn.ListenAndServe()
}

```

## License
[MIT](https://choosealicense.com/licenses/mit/)