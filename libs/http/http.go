package http

import (
	"errors"
	"io"
	"net/http"
	"team.gg-server/util"
)

type GetRequest struct {
	Url           string
	Authorization string
}

type PostRequest struct {
	Url           string
	Body          interface{}
	Authorization string
}

type Response struct {
	Success    bool
	StatusCode int
	Body       []byte
	Err        error
}

func Get(req GetRequest) (Response, error) {
	var respBody Response

	resp, err := http.Get(req.Url)
	if err != nil {
		return respBody, err
	}
	defer resp.Body.Close()

	bodyContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return respBody, err
	}

	respBody.StatusCode = resp.StatusCode
	respBody.Success = resp.StatusCode >= 200 && resp.StatusCode < 300
	if respBody.Success {
		respBody.Body = bodyContent
	} else {
		respBody.Err = errors.New(string(bodyContent))
	}
	return respBody, nil
}

func Post(req PostRequest) (Response, error) {
	var respBody Response

	requestBody := util.StructToReadable(req.Body)
	request, err := http.NewRequest("POST", req.Url, requestBody)
	if err != nil {
		return respBody, err
	}
	request.Header.Set("Content-Type", "application/json")
	if req.Authorization != "" {
		request.Header.Set("Authorization", "Bearer "+req.Authorization)
	}

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return respBody, err
	}
	defer resp.Body.Close()

	bodyContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return respBody, err
	}

	respBody.StatusCode = resp.StatusCode
	respBody.Success = resp.StatusCode >= 200 && resp.StatusCode < 300
	if respBody.Success {
		respBody.Body = bodyContent
	} else {
		respBody.Err = errors.New(string(bodyContent))
	}
	return respBody, nil
}
