package api

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"media-server/src/models"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type MediaAPI struct {
	Host             string
	Port             int
	Prefix           string
	RootPath         string
	DataStorageRoute string
}

func NewMediaAPI(config *models.MediaAPIConfig) (*MediaAPI, error) {
	api := &MediaAPI{Host: config.Host, Port: config.Port, Prefix: config.Prefix, RootPath: config.StorageRootPath, DataStorageRoute: config.DataStorageRoute}
	users = make(map[string]string)
	users["admin"] = config.AdminPass
	return api, nil
}

func (api *MediaAPI) getFileExtension(fileName string) string {
	fileExtension := "_"
	fileNameSplitted := (strings.Split(fileName, "."))
	if len(fileNameSplitted) > 0 {
		fileExtension = fileNameSplitted[len(fileNameSplitted)-1]
	}
	return fileExtension
}

func (api *MediaAPI) encodeFileName(fileName string, fileExtension string) string {
	encodedFileName := md5.Sum([]byte(fileName))
	return fmt.Sprintf("%s.%s", hex.EncodeToString(encodedFileName[:]), fileExtension)
}

func (api *MediaAPI) getFullFilePath(fileExtension string) string {
	path := fmt.Sprintf("%s/%s/%s", api.RootPath, api.DataStorageRoute, fileExtension)
	reg := regexp.MustCompile("(/)*")
	return reg.ReplaceAllString(path, "$1")
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
	http.HandleFunc(fmt.Sprintf("%s%s", api.Prefix, "/ping/"), api.ping)
	http.HandleFunc(fmt.Sprintf("%s%s", api.Prefix, "/data/"), api.getDataByUrl)
	http.HandleFunc(fmt.Sprintf("%s%s", api.Prefix, "/signin"), api.signIn)
	http.HandleFunc(fmt.Sprintf("%s%s", api.Prefix, "/upload"), api.upload)
	http.HandleFunc(fmt.Sprintf("%s%s", api.Prefix, "/delete"), api.delete)
	log.Printf("Run server on %s:%d", api.Host, api.Port)
	http.ListenAndServe(fmt.Sprintf("%s:%d", api.Host, api.Port), nil)
}

func (api *MediaAPI) ping(rw http.ResponseWriter, r *http.Request) {
	api.SetCorsHeaders(&rw)
	fmt.Fprintln(rw, "pong")
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
	fmt.Fprintf(rw, "token:\"%s\"", token)
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
	fileExtension := api.getFileExtension(handler.Filename)
	fileName := api.encodeFileName(handler.Filename, fileExtension)
	fullDataStorageDestination := api.getFullFilePath(fileExtension)
	if _, err := os.Stat(fmt.Sprintf("%s/%s", fullDataStorageDestination, fileName)); err == nil {
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rw, "json:\"%s\"", "File already exist!")
		return
	}
	defer fileBytes.Close()
	os.MkdirAll(fullDataStorageDestination, os.ModePerm)
	data, err := ioutil.ReadAll(fileBytes)
	file, err := os.Create(fmt.Sprintf("%s/%s", fullDataStorageDestination, fileName))
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = file.Write(data)
	if err != nil {
		log.Println(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()
	fmt.Fprint(rw, fileName)
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
	fileExtension := api.getFileExtension(fileName)
	fullDataStorageDestination := api.getFullFilePath(fileExtension)
	if os.Remove(fmt.Sprintf("%s/%s", fullDataStorageDestination, fileName)) != nil {
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(rw, "json:\"%s\"", "File not found!")
	}
	os.Remove(fullDataStorageDestination)
	return
}

func (api *MediaAPI) getDataByUrl(rw http.ResponseWriter, r *http.Request) {
	fileName := r.URL.RequestURI()[len(fmt.Sprintf("%s%s", api.Prefix, "/data/")):]
	fileExtension := api.getFileExtension(fileName)
	fullDataStorageDestination := api.getFullFilePath(fileExtension)
	http.ServeFile(rw, r, fmt.Sprintf("%s/%s", fullDataStorageDestination, fileName))
}
