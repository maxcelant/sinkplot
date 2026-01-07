package runtime

import "net/http"

func NewWorkerServer(dh http.Handler) http.Server {
	return http.Server{
		Addr:    ":8080",
		Handler: dh,
	}

}
