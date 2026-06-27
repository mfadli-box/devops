package main

import (
	"ict_rest/backbone"
)

func main() {
	rest := backbone.SetRouter()
	defer backbone.PgSQL.Close()
	rest.Run(":36665")
}
