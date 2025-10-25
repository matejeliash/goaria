package ariarpc

import "encoding/json"

// Basic RPC requenst used in JSON
type JsonRpcRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	ID      string        `json:"id"`
	Params  []interface{} `json:"params"`
}

// Used for message received after request
type JsonRpcResponse struct {
	ID     string          `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type FileInfo struct {
	Path            string `json:"path"`
	CompletedLength string `json:"completedLength"`
	Length          string `json:"length"`
}

// Stored data about specific download
type DownloadData struct {
	GID             string     `json:"gid"`
	Status          string     `json:"status"`
	TotalLength     string     `json:"totalLength"`
	CompletedLength string     `json:"completedLength"`
	DownloadSpeed   string     `json:"downloadSpeed"`
	Dir             string     `json:"dir"`
	Connections     string     `json:"connections"`
	Files           []FileInfo `json:"files"`
}
