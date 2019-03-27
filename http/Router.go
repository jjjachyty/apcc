package http

import "net/http"

func init() {

	http.HandleFunc("/wallet/blance", GetBlance)

}
