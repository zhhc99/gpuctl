package ipc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/zhhc99/gpuctl/internal/locale"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dial()
		},
	},
	Timeout: 30 * time.Second,
}

func post[Req any, Resp any](path string, req Req) (Resp, error) {
	var zero Resp
	body, err := json.Marshal(req)
	if err != nil {
		return zero, fmt.Errorf("IPC marshal: %w", err)
	}
	resp, err := httpClient.Post("http://gpud"+path, "application/json", bytes.NewReader(body))
	if err != nil {
		if !isServicePresent() {
			return zero, fmt.Errorf("%s", locale.T("err.service_not_running"))
		}
		return zero, fmt.Errorf("%s", locale.T("err.service_not_responding"))
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&zero); err != nil {
		return zero, fmt.Errorf("IPC decode: %w", err)
	}
	return zero, nil
}

func PostLoad() (LoadResponse, error) {
	return post[struct{}, LoadResponse]("/load", struct{}{})
}

func PostList(req ListRequest) (ListResponse, error) {
	return post[ListRequest, ListResponse]("/list", req)
}

func PostTuneGet(req TuneGetRequest) (TuneGetResponse, error) {
	return post[TuneGetRequest, TuneGetResponse]("/tune/get", req)
}

func PostTuneSet(req TuneSetRequest) (TuneSetResponse, error) {
	return post[TuneSetRequest, TuneSetResponse]("/tune/set", req)
}

func PostTuneReset(req TuneResetRequest) (TuneResetResponse, error) {
	return post[TuneResetRequest, TuneResetResponse]("/tune/reset", req)
}

func PostVersion() (VersionResponse, error) {
	return post[struct{}, VersionResponse]("/version", struct{}{})
}

// IsRunning returns true if the service socket is present and reachable.
func IsRunning() bool {
	if !isServicePresent() {
		return false
	}
	_, err := PostVersion()
	return err == nil
}
