package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/shahanmmiah/Chirpy/internal/database"
)

type Handler struct {
	Ns     string
	Handle http.Handler
	Method string
}

type Handlers struct {
	HandleData map[string]Handler
}

func RedinisHandler() http.Handler {

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
		resp.WriteHeader(200)

		stat, _ := resp.Write([]byte("OK"))

		fmt.Printf("%v - %v", resp.Header(), stat)
	})
}

type ApiConfig struct {
	fileserverHits atomic.Int32
	DbQueries      *database.Queries
}

func (a *ApiConfig) MiddlewareIncHits(handler http.Handler) http.Handler {

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		a.fileserverHits.Add(1)
		//fmt.Printf("incrmenting Hit to : %v", a.fileserverHits.Load())
		handler.ServeHTTP(resp, req)
	})

}

func (a *ApiConfig) MiddlewareReqCheckHandle() http.Handler {

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/html")
		resp.WriteHeader(200)
		resp.Write([]byte(fmt.Sprintf(
			`<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
		</html>`, a.fileserverHits.Load())))
	})

}

func (a *ApiConfig) MiddlewareReqResetHandle() http.Handler {

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		a.fileserverHits.Store(0)
		resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
		resp.WriteHeader(200)
		resp.Write([]byte(fmt.Sprintf("Server hits reset to: %v ", a.fileserverHits.Load())))
	})

}

func ErrorJsonResp(resp http.ResponseWriter, err error) {
	errData := struct {
		Error string `json:"error"`
	}{Error: fmt.Sprintf("error %v", err)}

	jsonData, _ := json.Marshal(errData)
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(400)
	resp.Write(jsonData)
}

func SanatizeProfane(text string) string {

	profaneWords := map[string]string{"kerfuffle": CENSORSTR, "sharbert": CENSORSTR, "fornax": CENSORSTR}
	cleanStr := text

	for _, word := range strings.Fields(text) {

		if _, found := profaneWords[strings.ToLower(word)]; found {
			cleanStr = strings.Replace(cleanStr, word, CENSORSTR, 1)
		}

	}
	return cleanStr
}

func (a *ApiConfig) MiddlewareValidateChirp(chripLen int) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resData := struct {
			Body string `json:"body"`
		}{}

		reqData, err := io.ReadAll(req.Body)

		if err != nil {
			ErrorJsonResp(resp, err)
			return

		}

		err = json.Unmarshal(reqData, &resData)

		if err != nil {
			ErrorJsonResp(resp, err)
			return
		}

		if len(resData.Body) > chripLen {
			ErrorJsonResp(resp, fmt.Errorf("error: Chirp is too long"))
			return
		}

		ChirpData := struct {
			Cleaned_body string `json:"cleaned_body"`
		}{Cleaned_body: SanatizeProfane(resData.Body)}

		jsonData, _ := json.Marshal(ChirpData)
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(200)
		resp.Write(jsonData)

	})

}

func (h *Handlers) HandleHandlers(mux *http.ServeMux) error {

	for hndlName, handle := range h.HandleData {

		if handle.Ns == "" {
			return fmt.Errorf("%v Ns cannot be empty", hndlName)
		}
		if handle.Method == "" {
			return fmt.Errorf("%v Method cannot be empty", hndlName)
		}

		mux.Handle(fmt.Sprintf("%s %s%s", handle.Method, handle.Ns, hndlName), handle.Handle)

	}
	return nil
}

func main() {
	godotenv.Load()

	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	a := ApiConfig{}
	a.DbQueries = database.New(db)

	handleMap := Handlers{HandleData: map[string]Handler{}}

	fileServeHandler := http.FileServer(http.Dir("."))
	//ApiServeHandler := http.

	handleMap.HandleData["/healthz"] = Handler{Ns: BACKEND_NS, Handle: RedinisHandler(), Method: GET_METHOD}
	handleMap.HandleData["/reset"] = Handler{Ns: ADMIN_NS, Handle: a.MiddlewareReqResetHandle(), Method: POST_METHOD}
	handleMap.HandleData["/metrics"] = Handler{Ns: ADMIN_NS, Handle: a.MiddlewareReqCheckHandle(), Method: GET_METHOD}
	handleMap.HandleData["/"] = Handler{Ns: FRONTEND_NS, Handle: a.MiddlewareIncHits(http.StripPrefix("/app", fileServeHandler)), Method: GET_METHOD}
	handleMap.HandleData["/validate_chirp"] = Handler{Ns: BACKEND_NS, Handle: a.MiddlewareValidateChirp(140), Method: POST_METHOD}

	handleMap.HandleHandlers(mux)

	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	server.ListenAndServe()

}
