package main

import (
	"net/http"

	checkout "github.com/jekiapp/hi-mod-arch/internal/usecase/checkout"
	handlerPkg "github.com/jekiapp/hi-mod-arch/pkg/handler"
)

type Handler struct {
	CheckoutPageHandler handlerPkg.GenericHandlerHttp[checkout.CheckoutPageRequest, checkout.CheckoutPageResponse]
}

func (h Handler) routes(mux *http.ServeMux) {
	mux.HandleFunc("/checkout", handlerPkg.HttpGenericHandler(h.CheckoutPageHandler))
}
