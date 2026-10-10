package main

import (
	"BE2_API_Shop/shop/cfg"
	"BE2_API_Shop/shop/web"
	"flag"
	"log"
	"net/http"
)

// @title Appliance Shop API
// @version 1.0
// @description REST API for household appliance shop
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter JWT token in format: Bearer <token>
//
// @host localhost:4000
// @BasePath /api/v1
func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")

	flag.Parse()
	conf, err := cfg.Load("shop/cfg/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	app := web.NewApp(conf)

	defer app.GrpcAuthConnClose()
	mux := app.Routers()

	httpAddr := conf.Shop.HTTPAddr
	if *addr != "" {
		httpAddr = *addr
	}

	log.Printf("starting server on %s", httpAddr)

	err = http.ListenAndServe(httpAddr, mux)
	if err != nil {
		log.Fatal(err)
	}

}
