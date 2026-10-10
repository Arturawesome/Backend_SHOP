package web

import (
	"net/http"

	_ "BE2_API_Shop/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func (app *Application) Routers() *http.ServeMux {
	mux := http.NewServeMux()

	// Auth
	fileServer := http.FileServer(http.Dir("./static"))

	mux.Handle(
		"GET /static/",
		http.StripPrefix("/static/", fileServer),
	)
	mux.HandleFunc("GET /api/v1/sse", app.sseHandler)
	mux.HandleFunc("GET /api/v1/websocket", app.webSocketHandler)

	mux.HandleFunc("POST /api/v1/register", app.Register)
	mux.HandleFunc("POST /api/v1/auth", app.Login)
	mux.HandleFunc("POST /api/v1/reset", app.ResetPassword)
	mux.Handle("PATCH /api/v1/auth/password",
		app.authJwtMiddleware(http.HandlerFunc(app.ChangePassword)))

	mux.Handle("POST /api/v1/clients",
		app.authJwtMiddleware(http.HandlerFunc(app.ClientAdd)))

	mux.Handle("GET /api/v1/clients",
		app.authJwtMiddleware(http.HandlerFunc(app.ClientList)))
	mux.Handle("GET /api/v1/clients/search",
		app.authJwtMiddleware(http.HandlerFunc(app.ClientSearch)))
	mux.Handle("DELETE /api/v1/clients/{id}",
		app.authJwtMiddleware(http.HandlerFunc(app.ClientRemove)))

	mux.Handle("PATCH /api/v1/clients/{id}/address",
		app.authJwtMiddleware(http.HandlerFunc(app.ClientPatchAddress)))

	mux.Handle("POST /api/v1/products",
		app.authJwtMiddleware(http.HandlerFunc(app.ProductAdd)))

	mux.Handle("GET /api/v1/products",
		app.authJwtMiddleware(http.HandlerFunc(app.ProductGetAvailable)))

	mux.Handle("GET /api/v1/products/{id}",
		app.authJwtMiddleware(http.HandlerFunc(app.ProductGetByID)))
	mux.Handle("PATCH /api/v1/products/{id}/stock",
		app.authJwtMiddleware(http.HandlerFunc(app.ProductDecrease)))
	mux.Handle("DELETE /api/v1/products/{id}",
		app.authJwtMiddleware(http.HandlerFunc(app.ProductRemoveByID)))

	mux.Handle("POST /api/v1/suppliers",
		app.authJwtMiddleware(http.HandlerFunc(app.SupplierAdd)))
	mux.Handle("GET /api/v1/suppliers",
		app.authJwtMiddleware(http.HandlerFunc(app.SupplierList)))
	mux.Handle("DELETE /api/v1/suppliers/{id}",
		app.authJwtMiddleware(http.HandlerFunc(app.SupplierRemoveByID)))
	mux.Handle("GET /api/v1/suppliers/{id}",
		app.authJwtMiddleware(http.HandlerFunc(app.SupplierGetByID)))
	mux.Handle("PATCH /api/v1/suppliers/{id}/address",
		app.authJwtMiddleware(http.HandlerFunc(app.SupplierPatchAddress)))

	mux.Handle("POST /api/v1/products/{id}/image",
		app.authJwtMiddleware(http.HandlerFunc(app.ImageAdd)))
	mux.Handle("GET /api/v1/products/{id}/image",
		app.authJwtMiddleware(http.HandlerFunc(app.ImageGetByProductID)))
	mux.Handle("DELETE /api/v1/images/{id}",
		app.authJwtMiddleware(http.HandlerFunc(app.ImageRemove)))
	mux.Handle("GET /api/v1/images/{id}",
		app.authJwtMiddleware(http.HandlerFunc(app.ImageGet)))
	mux.Handle("PUT /api/v1/images/{id}",
		app.authJwtMiddleware(http.HandlerFunc(app.ImageChange)))

	mux.Handle(
		"/swagger/",
		httpSwagger.WrapHandler,
	)
	return mux
}
