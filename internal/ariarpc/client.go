package ariarpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AriaClient struct {
	RpcSecret  string
	HttpClient *http.Client
	RpcUrl     string
}

func NewAriaClient(rpcSecret string) *AriaClient {
	ariaClient := &AriaClient{}

	ariaClient.HttpClient = &http.Client{
		Timeout: 3 * time.Second,
	}
	ariaClient.RpcUrl = "http://localhost:6800/jsonrpc"
	ariaClient.RpcSecret = rpcSecret

	return ariaClient
}

func (ac *AriaClient) CreateSingleMethodRequest(method string, params []any) *JsonRpcRequest {
	reqBody := &JsonRpcRequest{
		Jsonrpc: "2.0",
		Method:  method,
		ID:      "goaria",
		Params:  append([]any{"token:" + ac.RpcSecret}, params...),
	}
	return reqBody

}

func (ac *AriaClient) ShutdownAriaProcess() error {

	req := ac.CreateSingleMethodRequest("aria2.shutdown", []any{})
	_, err := ac.CallJsonRpc(req)
	if err != nil {
		return err
	}
	return nil

}

func (ac *AriaClient) PauseDownload(gid string) (*JsonRpcResponse, error) {

	req := &JsonRpcRequest{
		Jsonrpc: "2.0",
		Method:  "aria2.pause",
		ID:      "goaria",
		Params:  []any{"token:" + ac.RpcSecret, gid},
	}

	jsonRpcResp, err := ac.CallJsonRpc(req)
	if err != nil {
		return nil, err
	}
	return jsonRpcResp, nil

}

func (ac *AriaClient) RemoveDownload(gid string) (*JsonRpcResponse, error) {

	req := &JsonRpcRequest{
		Jsonrpc: "2.0",
		Method:  "aria2.remove",
		ID:      "goaria",
		Params:  []any{"token:" + ac.RpcSecret, gid},
	}

	jsonRpcResp, err := ac.CallJsonRpc(req)
	if err != nil {
		return nil, err
	}
	return jsonRpcResp, nil

}

func (ac *AriaClient) UnpauseDownload(gid string) (*JsonRpcResponse, error) {

	req := &JsonRpcRequest{
		Jsonrpc: "2.0",
		Method:  "aria2.unpause",
		ID:      "goaria",
		Params:  []any{"token:" + ac.RpcSecret, gid},
	}

	jsonRpcResp, err := ac.CallJsonRpc(req)
	if err != nil {
		return nil, err
	}
	return jsonRpcResp, nil

}

func (ac *AriaClient) CreateTellActiveReq() *JsonRpcRequest {
	return &JsonRpcRequest{
		Jsonrpc: "2.0",
		Method:  "aria2.tellActive",
		ID:      "active-downloads",
		Params:  []any{"token:" + ac.RpcSecret}, // no extra params needed
	}
}

func (ac *AriaClient) CreateTellWaitingReq() *JsonRpcRequest {
	return &JsonRpcRequest{
		Jsonrpc: "2.0",
		Method:  "aria2.tellWaiting",
		ID:      "active-downloads",
		Params:  []any{"token:" + ac.RpcSecret, 0, 1000},
	}
}

func (ac *AriaClient) CallJsonRpc(jsonRpcRequest *JsonRpcRequest) (*JsonRpcResponse, error) {

	data, err := json.Marshal(jsonRpcRequest)
	if err != nil {
		return nil, err
	}

	// Create a new POST request
	req, err := http.NewRequest("POST", ac.RpcUrl, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Reuse the client
	resp, err := ac.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result JsonRpcResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("FetchData: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("aria2 error %d: %s", result.Error.Code, result.Error.Message)
	}

	return &result, nil
}

// Returns all active and paused downloads
func (ac *AriaClient) GetRelevantDownloads() ([]DownloadData, error) {
	jsonRpcRequest := ac.CreateTellActiveReq()

	result, err := ac.CallJsonRpc(jsonRpcRequest)
	if err != nil {
		return nil, err
	}

	var active []DownloadData
	err = json.Unmarshal(result.Result, &active)
	if err != nil {
		return nil, err
	}

	getPauseRequest := ac.CreateTellWaitingReq()

	result2, err := ac.CallJsonRpc(getPauseRequest)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var paused []DownloadData
	err = json.Unmarshal(result2.Result, &paused)
	if err != nil {
		return nil, err
	}
	for _, p := range paused {
		active = append(active, p)
	}

	return active, nil
}

func (ac *AriaClient) AddDownload(url, filename, dir string) (*JsonRpcResponse, error) {

	var params []any // all params

	var err error


	if err != nil {
		return nil, err
	}
	uris := []string{url}      // list of URIs
	var options map[string]any // options inside params

	if filename != "" && dir != "" {
		options = map[string]any{ // options map
			"out": filename, // force the download filename
			"dir": dir,
		}
		params = []any{uris, options}

	} else if filename != "" {
		options = map[string]any{ // options map
			"out": filename, // force the download filename
		}
		params = []any{uris, options}

	} else if dir != "" {
		options = map[string]any{ // options map
			"dir": dir,
		}
		params = []any{uris, options}

	} else {
		params = []any{uris}
	}

	fmt.Println(params)

	// Add the URL to aria2
	jsonRpcReq := ac.CreateSingleMethodRequest("aria2.addUri", params)
	result, err := ac.CallJsonRpc(jsonRpcReq)
	if err != nil {
		return nil, err
	}

	return result, nil

}
