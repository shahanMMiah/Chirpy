package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/shahanmmiah/Chirpy/internal/auth"
	"github.com/shahanmmiah/Chirpy/internal/database"
)

type Handler struct {
	Ns     string
	Handle http.Handler
}

type Handlers map[string]map[string]Handler

type ChirpJson struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type UserJson struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type UserDbJson struct {
	ID         uuid.UUID `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdateddAt time.Time `json:"updated_at"`
	Email      string    `json:"email"`
}

func RedinisHandler() http.Handler {

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
		resp.WriteHeader(OKCODE)

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
		resp.WriteHeader(OKCODE)
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

		a.MiddleWareResetUsers().ServeHTTP(resp, req)
		a.MiddleWareResetChirps().ServeHTTP(resp, req)

		resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
		resp.WriteHeader(OKCODE)
		resp.Write([]byte(fmt.Sprintf("Server hits reset to: %v\n ", a.fileserverHits.Load())))
	})

}

func ErrorJsonResp(resp http.ResponseWriter, err error, errorCode int) {
	errData := struct {
		Error string `json:"error"`
	}{Error: fmt.Sprintf("error %v", err)}

	jsonData, _ := json.Marshal(errData)
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(errorCode)
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

func ValidateChirp(chirp string, chripLen int) bool {
	return len(chirp) <= chripLen
}

func (a *ApiConfig) MiddlewareGetChirps(id uuid.UUID) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {

		chirpDb, err := a.DbQueries.GetChirps(req.Context(), id)
		if err != nil {
			ErrorJsonResp(resp, err, FAILEDCODE)
		}

		chirpJson := ChirpJson{
			ID:        chirpDb.ID,
			CreatedAt: chirpDb.CreatedAt,
			UpdatedAt: chirpDb.UpdatedAt,
			Body:      chirpDb.Body,
			UserID:    chirpDb.UserID}

		jsonData, err := json.Marshal(chirpJson)
		if err != nil {
			ErrorJsonResp(resp, err, FAILEDCODE)
			return

		}

		resp.Header().Set("Content-type", "application/json")
		resp.WriteHeader(OKCODE)
		resp.Write(jsonData)

	})

}

func (a *ApiConfig) MiddlewareGetAllChirps() http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		chirpsDb, err := a.DbQueries.GetAllChirps(req.Context())
		if err != nil {
			ErrorJsonResp(resp, err, FAILEDCODE)
			return

		}

		chirpsJson := []ChirpJson{}

		for _, c := range chirpsDb {
			chirpsJson = append(chirpsJson, ChirpJson{
				ID:        c.ID,
				CreatedAt: c.CreatedAt,
				UpdatedAt: c.UpdatedAt,
				Body:      c.Body,
				UserID:    c.UserID})
		}

		slices.SortFunc(chirpsJson, func(a, b ChirpJson) int {
			return a.CreatedAt.Compare(b.CreatedAt)
		})

		jsonData, err := json.Marshal(chirpsJson)
		if err != nil {
			ErrorJsonResp(resp, err, FAILEDCODE)
			return

		}

		resp.Header().Set("Content-type", "application/json")
		resp.WriteHeader(OKCODE)
		resp.Write(jsonData)
	})
}

func (a *ApiConfig) MiddlewareAddChirp(chirpLen int, mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resData := struct {
			Body   string `json:"body"`
			UserId string `json:"user_id"`
		}{}

		reqData, err := io.ReadAll(req.Body)

		if err != nil {
			ErrorJsonResp(resp, err, FAILEDCODE)
			return

		}

		err = json.Unmarshal(reqData, &resData)

		if err != nil {
			ErrorJsonResp(resp, err, FAILEDCODE)
			return
		}

		if !ValidateChirp(resData.Body, chirpLen) {
			ErrorJsonResp(resp, fmt.Errorf("error: Chirp is too long"), FAILEDCODE)
			return
		}

		userId, err := uuid.Parse(resData.UserId)
		if err != nil {
			ErrorJsonResp(resp, err, FAILEDCODE)
			return
		}
		chirpDbData, err := a.DbQueries.CreateChirps(req.Context(), database.CreateChirpsParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Body:      SanatizeProfane(resData.Body),
			UserID:    userId})

		if err != nil {
			ErrorJsonResp(resp, err, FAILEDCODE)
			return
		}

		err = HandleHandler(
			mux,
			Handler{Ns: BACKEND_NS, Handle: a.MiddlewareGetChirps(chirpDbData.ID)},
			fmt.Sprintf("/chirps/%s", chirpDbData.ID.String()),
			GET_METHOD)

		if err != nil {
			ErrorJsonResp(resp, err, FAILEDCODE)
			return
		}

		ChirpData := struct {
			ID        uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Body      string    `json:"body"`
			UserID    uuid.UUID `json:"user_id"`
		}{
			ID:        chirpDbData.ID,
			CreatedAt: chirpDbData.CreatedAt,
			UpdatedAt: chirpDbData.UpdatedAt,
			Body:      chirpDbData.Body,
			UserID:    chirpDbData.UserID}

		jsonData, _ := json.Marshal(ChirpData)
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(NEWCODE)
		resp.Write(jsonData)

	})

}

func (a *ApiConfig) MiddleWareCreateUserHandle() http.Handler {

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {

		emailStruct := &UserJson{}
		reqData, err := io.ReadAll(req.Body)
		if err != nil {
			ErrorJsonResp(resp, err, FAILEDCODE)
		}

		err = json.Unmarshal(reqData, emailStruct)
		if err != nil {
			ErrorJsonResp(resp, err, FAILEDCODE)
		}

		HashedPassword, err := auth.HashPassword(emailStruct.Password)
		if err != nil {
			ErrorJsonResp(resp, err, FAILEDCODE)
		}

		params := database.CreateUserParams{
			ID:             uuid.New(),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			Email:          emailStruct.Email,
			HashedPassword: HashedPassword,
		}
		userDbQuiery, err := a.DbQueries.CreateUser(req.Context(), params)
		if err != nil {
			ErrorJsonResp(resp, err, FAILEDCODE)
		}

		userDbStruct := UserDbJson{
			ID:         userDbQuiery.ID,
			CreatedAt:  userDbQuiery.CreatedAt,
			UpdateddAt: userDbQuiery.CreatedAt,
			Email:      userDbQuiery.Email}

		userData, err := json.Marshal(userDbStruct)
		if err != nil {
			ErrorJsonResp(resp, err, FAILEDCODE)
		}

		resp.Header().Set("Content-Type", "application:json")
		resp.WriteHeader(NEWCODE)
		resp.Write(userData)

	})
}

func (a *ApiConfig) MiddleWareResetUsers() http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if os.Getenv("PLATFORM") != "dev" {
			ErrorJsonResp(resp, fmt.Errorf("Forbidden"), FORBIDDENCODE)
		}

		a.DbQueries.ResetUsers(req.Context())
		resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
		resp.WriteHeader(OKCODE)
		resp.Write([]byte("users database has been reset\n"))
	})
}

func (a *ApiConfig) MiddleWareResetChirps() http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if os.Getenv("PLATFORM") != "dev" {
			ErrorJsonResp(resp, fmt.Errorf("Forbidden"), FORBIDDENCODE)
		}

		a.DbQueries.ResetChirps(req.Context())
		resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
		resp.WriteHeader(OKCODE)
		resp.Write([]byte("chirps database has been reset\n"))
	})
}

func (a *ApiConfig) MiddlewareLoginHandler() http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		userJson := &UserJson{}

		reqData, err := io.ReadAll(req.Body)
		if err != nil {
			ErrorJsonResp(resp, err, FAILEDCODE)
		}

		err = json.Unmarshal(reqData, userJson)
		if err != nil {
			ErrorJsonResp(resp, err, FAILEDCODE)
		}

		userDb, err := a.DbQueries.GetUserFromEmail(req.Context(), userJson.Email)
		if err != nil {
			ErrorJsonResp(resp, err, FAILEDCODE)
		}

		err = auth.CheckPasswordHash(userJson.Password, userDb.HashedPassword)
		if err != nil {
			ErrorJsonResp(resp, err, UNAUTHORIZED)
		}

		userDbjson := UserDbJson{ID: userDb.ID,
			CreatedAt:  userDb.CreatedAt,
			UpdateddAt: userDb.CreatedAt,
			Email:      userDb.Email}

		userDbjsonData, err := json.Marshal(userDbjson)
		if err != nil {
			ErrorJsonResp(resp, err, FAILEDCODE)
		}

		resp.Header().Set("Content-type", "application:json")
		resp.WriteHeader(OKCODE)
		resp.Write(userDbjsonData)

	})
}

func HandleHandler(mux *http.ServeMux, handle Handler, hndlName, mthdName string) error {

	if handle.Ns == "" {
		return fmt.Errorf("%v Ns cannot be empty", hndlName)
	}
	if mthdName == "" {
		return fmt.Errorf("%v Method cannot be empty", hndlName)
	}

	mux.Handle(fmt.Sprintf("%s %s%s", mthdName, handle.Ns, hndlName), handle.Handle)

	return nil
}

func HandleHandlers(mux *http.ServeMux, h Handlers) error {

	for hndlName, mthds := range h {
		for mthdName, handle := range mthds {
			err := HandleHandler(mux, handle, hndlName, mthdName)
			if err != nil {
				return err
			}

		}
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

	type handlerMap map[string]Handler
	endpointMap := Handlers{}

	fileServeHandler := http.FileServer(http.Dir("."))
	//ApiServeHandler := http.

	// admin handlers
	endpointMap["/reset"] = handlerMap{POST_METHOD: Handler{Ns: ADMIN_NS, Handle: a.MiddlewareReqResetHandle()}}
	endpointMap["/metrics"] = handlerMap{GET_METHOD: Handler{Ns: ADMIN_NS, Handle: a.MiddlewareReqCheckHandle()}}

	// api handlers
	endpointMap["/chirps"] = handlerMap{
		POST_METHOD: Handler{Ns: BACKEND_NS, Handle: a.MiddlewareAddChirp(140, mux)},
		GET_METHOD:  Handler{Ns: BACKEND_NS, Handle: a.MiddlewareGetAllChirps()}}

	endpointMap["/users"] = handlerMap{POST_METHOD: Handler{Ns: BACKEND_NS, Handle: a.MiddleWareCreateUserHandle()}}
	endpointMap["/healthz"] = handlerMap{POST_METHOD: Handler{Ns: BACKEND_NS, Handle: RedinisHandler()}}
	endpointMap["/login"] = handlerMap{POST_METHOD: Handler{Ns: BACKEND_NS, Handle: a.MiddlewareLoginHandler()}}

	// frontend handlers
	endpointMap["/"] = handlerMap{GET_METHOD: Handler{Ns: FRONTEND_NS, Handle: a.MiddlewareIncHits(http.StripPrefix("/app", fileServeHandler))}}

	HandleHandlers(mux, endpointMap)

	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	server.ListenAndServe()

}
