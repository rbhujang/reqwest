package reqwest

import "net/http"

type Middleware func(*http.Request) error
