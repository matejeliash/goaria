package handlers

import (
	"encoding/json"
	"fmt"
	"goaria/internal/ariarpc"
	"goaria/internal/session"
	"log"
	"net/http"
	"strconv"
)

type HandlerManager struct {
	AriaClient *ariarpc.AriaClient
}

func NewHandlerManager(ariaClient *ariarpc.AriaClient) *HandlerManager {
	hm := &HandlerManager{
		AriaClient: ariaClient,
	}
	return hm
}

func (hm *HandlerManager) RemoveDownloadHandler(w http.ResponseWriter, r *http.Request) {

	gid := r.PathValue("gid") // gid identified downlaod

	//send request
	jsonRpcResp, err := hm.AriaClient.RemoveDownload(gid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	responsePayload := struct {
		Status json.RawMessage `json:"status"` // Holds the Aria2 result, typically "OK"
		Gid    string          `json:"gid"`
	}{
		Status: jsonRpcResp.Result,
		Gid:    gid,
	}

	if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
		log.Printf("Error encoding response payload: %v", err)
	}

}

func (hm *HandlerManager) PauseDownloadHandler(w http.ResponseWriter, r *http.Request) {
	gid := r.PathValue("gid")

	log.Println(gid)

	jsonRpcResp, err := hm.AriaClient.PauseDownload(gid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the GID of the new download
	w.Header().Set("Content-Type", "application/json")

	responsePayload := struct {
		Status json.RawMessage `json:"status"` // Holds the Aria2 result, typically "OK"
		Gid    string          `json:"gid"`
	}{
		Status: jsonRpcResp.Result,
		Gid:    gid,
	}

	if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
		log.Printf("Error encoding response payload: %v", err)
		// Note: Cannot call http.Error here as headers were already set/written
	}

}

func (hm *HandlerManager) UnpauseDownloadHandler(w http.ResponseWriter, r *http.Request) {
	gid := r.PathValue("gid")

	log.Println(gid)

	jsonRpcResp, err := hm.AriaClient.UnpauseDownload(gid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the GID of the new download
	w.Header().Set("Content-Type", "application/json")

	responsePayload := struct {
		Status json.RawMessage `json:"status"` // Holds the Aria2 result, typically "OK"
		Gid    string          `json:"gid"`
	}{
		Status: jsonRpcResp.Result,
		Gid:    gid,
	}

	if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
		log.Printf("Error encoding response payload: %v", err)
	}

}

func (hm *HandlerManager) ActiveDownloadsHandler(w http.ResponseWriter, r *http.Request) {
	downloads, err := hm.AriaClient.GetRelevantDownloads()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// just not simple conversion from bytes to MB
	convertData := func(bytes string) string {
		b, err := strconv.ParseInt(bytes, 10, 64)
		if err != nil {
			return "0"
		}
		return fmt.Sprintf("%d MB", b/1024/1024)
	}

	for i, _ := range downloads {
		downloads[i].CompletedLength = convertData(downloads[i].CompletedLength)
		downloads[i].TotalLength = convertData(downloads[i].TotalLength)
		downloads[i].DownloadSpeed = convertData(downloads[i].DownloadSpeed) + "/s"
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(downloads); err != nil {
		log.Printf("Error encoding response payload: %v", err)
	}

}

func (hm *HandlerManager) LoginHandler(w http.ResponseWriter, r *http.Request) {
	store := session.GetStore()
	session, _ := store.Get(r, session.GetSessionName())

	// If already logged in, skip password check
	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Already logged in")
		return
	}

	// Only allow POST , non needed got it covered in server.go
	// if r.Method != http.MethodPost {
	// 	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	// 	return
	// }

	// Get password from form
	password := r.FormValue("password")
	if password == "" {
		http.Error(w, "Password required", http.StatusBadRequest)
		return
	}

	// password is kept for testing, later there will be switch probably to ENV varaible
	if password != "password" {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// Set session as authenticated
	session.Values["authenticated"] = true
	session.Save(r, w)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Login successful")
}

func (hm *HandlerManager) LogoutHandler(w http.ResponseWriter, r *http.Request) {

	store := session.GetStore()
	session, _ := store.Get(r, session.GetSessionName())
	session.Values["authenticated"] = false
	session.Options.MaxAge = -1 // delete cookie
	session.Save(r, w)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Logged out")
}

func (hm *HandlerManager) DownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Bad http method", http.StatusBadRequest)
		return

	}
	// Parse the form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	url := r.FormValue("url")
	if url == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	dir := r.FormValue("dir")
	filename := r.FormValue("filename")
	fmt.Println(filename)
	fmt.Println(url)
	fmt.Println(dir)

	jsonRpcResp, err := hm.AriaClient.AddDownload(url, filename, dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responsePayload := struct {
		Status json.RawMessage `json:"status"` // Holds the Aria2 result, typically "OK"
		Gid    string          `json:"gid"`
	}{
		Status: jsonRpcResp.Result,
		Gid:    "",
	}

	// Respond with the GID of the new download
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(responsePayload); err != nil {
		log.Printf("Error encoding response payload: %v", err)
	}
}
