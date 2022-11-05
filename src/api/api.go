package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"media-server/src/models"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
)

type MediaAPI struct {
	Host     string
	Port     int
	Prefix   string
	RootPath string
	Routes   models.MediaAPIRoutes
}

func NewMediaAPI(config *models.MediaAPIConfig) (*MediaAPI, error) {
	api := &MediaAPI{Host: config.Host, Port: config.Port, Prefix: config.Prefix, RootPath: config.StorageRootPath, Routes: config.Routes}
	users = make(map[string]string)
	users["admin"] = config.AdminPass
	return api, nil
}

func (api *MediaAPI) setRapidHeader(r *http.Request) {
	r.Header.Add("content-type", "application/x-www-form-urlencoded")
	r.Header.Add("Accept-Encoding", "application/gzip")
	r.Header.Add("X-RapidAPI-Key", "1ae2aa72f1mshb729e169c8532bfp14534cjsn7ddb25386cc2")
	r.Header.Add("X-RapidAPI-Host", "google-translate1.p.rapidapi.com")

}

func (api *MediaAPI) authorization(r *http.Request) int {
	token := r.Header.Get("Token")
	if token == "" {
		return http.StatusUnauthorized
	}
	var claims models.Claims
	tokenValidation, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) { return jwtKey, nil })
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return http.StatusUnauthorized
		}
		return http.StatusBadRequest
	}
	if !tokenValidation.Valid {
		return http.StatusUnauthorized
	}
	return http.StatusOK
}

func (api *MediaAPI) SetCorsHeaders(rw *http.ResponseWriter) {
	(*rw).Header().Set("Access-Control-Allow-Origin", "*")
	(*rw).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (api *MediaAPI) Run() {
	http.HandleFunc(fmt.Sprintf("%s%s", api.Prefix, api.Routes.DataRoute.Name), api.getDataByUrl)
	http.HandleFunc(fmt.Sprintf("%s%s", api.Prefix, "/signin"), api.signIn)
	http.HandleFunc(fmt.Sprintf("%s%s", api.Prefix, "/upload"), api.upload)
	http.HandleFunc(fmt.Sprintf("%s%s", api.Prefix, "/delete"), api.delete)
	log.Printf("Run server on %s:%d", api.Host, api.Port)
	http.ListenAndServe(fmt.Sprintf("%s:%d", api.Host, api.Port), nil)
}

func (api *MediaAPI) signIn(rw http.ResponseWriter, r *http.Request) {
	var creds models.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	if !CredentialsValidation(creds) {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}
	token, err := GenerateNewToken(creds, 2)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Fprintf(rw, fmt.Sprintf("{\"token\": \"%s\"}", token))
}

func (api *MediaAPI) upload(rw http.ResponseWriter, r *http.Request) {
	code := api.authorization(r)
	if code != http.StatusOK {
		rw.WriteHeader(code)
		return
	}
	if r.Method != http.MethodPost {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	fileBytes, handler, err := r.FormFile("file")
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	if _, err := os.Stat(fmt.Sprintf("%s%s%s", api.RootPath, api.Routes.DataRoute.StorageRoute, handler.Filename)); err == nil {
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rw, "{\"error\": \"File already exist!\"}")
		return
	}
	defer fileBytes.Close()
	data, err := ioutil.ReadAll(fileBytes)
	file, _ := os.Create(fmt.Sprintf("%s%s%s", api.RootPath, api.Routes.DataRoute.StorageRoute, handler.Filename))
	file.Write(data)
	file.Close()
	log.Printf("%s%s%s", api.RootPath, api.Routes.DataRoute.StorageRoute, handler.Filename)
	stats, err := os.Stat(fmt.Sprintf("%s%s%s", api.RootPath, api.Routes.DataRoute.StorageRoute, handler.Filename))
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(rw, "{\"name\": %s\"}", stats.Name())
}

func (api *MediaAPI) delete(rw http.ResponseWriter, r *http.Request) {
	code := api.authorization(r)
	if code != http.StatusOK {
		rw.WriteHeader(code)
		return
	}
	if r.Method != http.MethodGet {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	fileName := r.URL.Query().Get("name")
	if _, err := os.Stat(fmt.Sprintf("%s%s%s", api.RootPath, api.Routes.DataRoute.StorageRoute, fileName)); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rw, "{\"error\": \"File not found!\"}")
	}
	os.Remove(fmt.Sprintf("%s%s%s", api.RootPath, api.Routes.DataRoute.StorageRoute, fileName))
	return
}

func (api *MediaAPI) getDataByUrl(rw http.ResponseWriter, r *http.Request) {
	currentPath := r.URL.RequestURI()[len(fmt.Sprintf("%s%s", api.Prefix, api.Routes.DataRoute.Name)):]
	http.ServeFile(rw, r, fmt.Sprintf("%s%s%s", api.RootPath, api.Routes.DataRoute.StorageRoute, currentPath))
}
