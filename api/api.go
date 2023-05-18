package api

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/EugeneKrivoshein/wb_l0/internal/db"

	"github.com/go-chi/chi/v5"
)

type ordkey string

const orderKey ordkey = "order"

type Api struct {
	rtr                *chi.Mux
	csh                *db.Cache
	name               string
	srv                *http.Server
	httpServerExitDone *sync.WaitGroup
}

func NewApi(csh *db.Cache) *Api {
	api := Api{}
	api.Init(csh)
	return &api
}

func (a *Api) Init(csh *db.Cache) {
	a.csh = csh
	a.name = "API"
	a.rtr = chi.NewRouter()
	a.rtr.Get("/", a.WellcomeHandler)

	a.rtr.Route("/orders", func(r chi.Router) {
		r.Route("/{orderID}", func(r chi.Router) {
			r.Use(a.orderCtx)
			r.Get("/", a.GetOrder)
		})
	})

	a.httpServerExitDone = &sync.WaitGroup{}
	a.httpServerExitDone.Add(1)
	a.StartServer()
}

func (a *Api) Finish() {
	log.Printf("%v: server shutdown...\n", a.name)

	if err := a.srv.Shutdown(context.Background()); err != nil {
		panic(err)
	}

	a.httpServerExitDone.Wait()
	log.Printf("%v: server stopped!\n", a.name)
}

func (a *Api) StartServer() {
	a.srv = &http.Server{
		Addr:    ":8080",
		Handler: a.rtr,
	}

	go func() {
		defer a.httpServerExitDone.Done()

		log.Printf("%v: server starting http://localhost:8080\n", a.name)
		if err := a.srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("ListenAndServe() error: %v", err)
			return
		}
	}()
}

func (a *Api) orderCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		orderIDstr := chi.URLParam(r, "orderID")
		orderID, err := strconv.ParseInt(orderIDstr, 10, 64)
		if err != nil {
			log.Printf("%v: err conv %s in int: %v\n", a.name, orderIDstr, err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		log.Printf("%v: query order from cache/db: %v\n", a.name, orderIDstr)
		orderOut, err := a.csh.GetOrderOutById(orderID)
		if err != nil {
			log.Printf("%v: err getting order from db: %v\n", a.name, err)
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		ctx := context.WithValue(r.Context(), orderKey, orderOut)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Api) WellcomeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("view/index.html")
	if err != nil {
		log.Printf("%v: err parsing html: %s\n", a.name, err)
		http.Error(w, "error", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = t.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Printf("%v: err parsing html: %s\n", a.name, err)
		return
	}
}

func (a *Api) GetOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orderOut, ok := ctx.Value(orderKey).(*db.OrderOut)
	if !ok {
		log.Printf("%v: error casting interface\n", a.name)
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	t, err := template.ParseFiles("view/index.html")
	if err != nil {
		log.Printf("%v: err parsing html: %s\n", a.name, err)
		http.Error(w, "error", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	t.ExecuteTemplate(w, "index.html", orderOut)
	if err != nil {
		log.Printf("%v: err parsing html: %s\n", a.name, err)
		return
	}
}
