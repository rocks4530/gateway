package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	gw "gateway/generated"
)

var (
	endpoint      = flag.String("endpoint", "grpc-server.apps.internal:6565", "Server EndPoint")
	swaggerpath   = "swaggerui"
	swaggerprefix = "/swaggerui/"
	port          = "8080"
	//os.Getenv("PORT")
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := gw.RegisterUserHandlerFromEndpoint(ctx, mux, *endpoint, opts)
	if err != nil {
		return err
	}
	swaggerMux := http.NewServeMux()
	serveSwagger(swaggerMux)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: grpcHandlerFunc(mux, swaggerMux),
	}
	fmt.Println("Started reverse proxy server, listening at " + port)
	return srv.ListenAndServe()
}
func serveSwagger(mux *http.ServeMux) {
	fs := http.FileServer(http.Dir("swaggerui"))
	mux.Handle(swaggerprefix, http.StripPrefix(swaggerprefix, fs))
}

func grpcHandlerFunc(ghandler http.Handler, shandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.RequestURI, swaggerpath) {
			shandler.ServeHTTP(w, r)
		} else {
			ghandler.ServeHTTP(w, r)
		}
	})
}

func main() {
	fmt.Println("Starting the reverse proxy server ....")
	flag.Parse()
	defer glog.Flush()
	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
