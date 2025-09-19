package main

import (
	"cleanArch_with_postgres/internal/infrastructure/app"
	"cleanArch_with_postgres/internal/infrastructure/router"
)

func main() {
	r := router.NewRouter()
	a := app.New(r)
	a.Start()
}
