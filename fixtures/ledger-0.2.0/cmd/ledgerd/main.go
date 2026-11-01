// Command ledgerd serves the in-memory ledger over HTTP.
package main

import (
	"flag"
	"log"
	"net/http"

	"ledger/internal/api"
	"ledger/internal/store"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "listen address")
	flag.Parse()

	st := store.New()
	st.CreateAccount("cash", "Cash")
	st.CreateAccount("bank", "Bank")

	srv := api.New(st)
	log.Printf("ledgerd listening on %s", *addr)
	if err := http.ListenAndServe(*addr, srv.Routes()); err != nil {
		log.Fatal(err)
	}
}
