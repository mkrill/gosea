package status

import (
	"fmt"
	"net/http"
	"time"
)

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain")

	//output := strings.NewReader("OK\n")
	//io.Copy(w, output)

	//w.Write([]byte("OK\n"))

	_, _ = fmt.Fprint(w, "Aktuelle Zeit: %v", time.Now().String())
}
