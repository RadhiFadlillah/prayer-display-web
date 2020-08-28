// +build !prod

package backend

import (
	"net/http"
)

var assets = http.Dir("internal/view")
