package generate

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/jsaund/gorest/parse"
	"github.com/stretchr/testify/assert"
)

func TestGenerateValid(t *testing.T) {
	src := `package test
		// @GET("/photos/{id}")
		type GetPhotoDetailsRequestBuilder interface {
			// @PATH("id")
			PhotoID(id string) GetPhotoDetailsRequestBuilder

			// @QUERY("image_size")
			ImageSize(size int) GetPhotoDetailsRequestBuilder

			// @SYNC("GetPhotoDetailsResponse")
			Run() (GetPhotoDetailsResponse, error)

			// @ASYNC("GetPhotoDetailsCallback")
			RunAsync(callback GetPhotoDetailsCallback)
		}
		`
	output := `/*
* CODE GENERATED AUTOMATICALLY WITH GOREST (github.com/jsaund/gorest)
* THIS FILE SHOULD NOT BE EDITED BY HAND
 */

package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/jsaund/gorest/restclient"
)

type GetPhotoDetailsCallback interface {
	OnStart()
	OnError(reason string)
	OnSuccess(response GetPhotoDetailsResponse)
}

type GetPhotoDetailsRequestBuilderImpl struct {
	pathSubstitutions  map[string]string
	queryParams        url.Values
	postFormParams     url.Values
	postBody           interface{}
	postMultiPartParam map[string][]byte
	headerParams       map[string]string
}

func NewGetPhotoDetailsRequestBuilder() GetPhotoDetailsRequestBuilder {
	return &GetPhotoDetailsRequestBuilderImpl{
		pathSubstitutions:  make(map[string]string),
		queryParams:        url.Values{},
		postFormParams:     url.Values{},
		postMultiPartParam: make(map[string][]byte),
		headerParams:       make(map[string]string),
	}
}

func (b *GetPhotoDetailsRequestBuilderImpl) PhotoID(id string) GetPhotoDetailsRequestBuilder {
	b.pathSubstitutions["id"] = fmt.Sprintf("%v", id)
	return b
}

func (b *GetPhotoDetailsRequestBuilderImpl) ImageSize(size int) GetPhotoDetailsRequestBuilder {
	b.queryParams.Add("image_size", fmt.Sprintf("%v", size))
	return b
}

func (b *GetPhotoDetailsRequestBuilderImpl) applyPathSubstituions(api string) string {
	if len(b.pathSubstitutions) == 0 {
		return api
	}

	for key, value := range b.pathSubstitutions {
		api = strings.Replace(api, "{"+key+"}", value, -1)
	}

	return api
}

func (b *GetPhotoDetailsRequestBuilderImpl) build() (req *http.Request, err error) {
	restClient := restclient.GetClient()
	if restClient == nil {
		return nil, fmt.Errorf("A rest client has not been registered yet. You must call client.RegisterClient first")
	}
	url := restClient.BaseURL() + b.applyPathSubstituions("/photos/{id}")
	httpMethod := "GET"
	switch httpMethod {
	case "POST", "PUT":
		if b.postBody != nil {
			// Assume the body is to be marshalled to JSON
			contentBody, err := json.Marshal(b.postBody)
			if err != nil {
				return nil, err
			}
			contentReader := bytes.NewReader(contentBody)
			req, err = http.NewRequest(httpMethod, url, contentReader)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Content-Type", "application/json")
		} else if len(b.postFormParams) > 0 {
			contentForm := b.postFormParams.Encode()
			contentReader := strings.NewReader(contentForm)
			if req, err = http.NewRequest(httpMethod, url, contentReader); err != nil {
				return nil, err
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		} else if len(b.postMultiPartParam) > 0 {
			contentBody := &bytes.Buffer{}
			writer := multipart.NewWriter(contentBody)
			for key, value := range b.postMultiPartParam {
				if err := writer.WriteField(key, string(value)); err != nil {
					return nil, err
				}
			}
			if err = writer.Close(); err != nil {
				return nil, err
			}
			if req, err = http.NewRequest(httpMethod, url, contentBody); err != nil {
				return nil, err
			}
			req.Header.Set("Content-Type", "multipart/form-data")
		}
	case "GET", "DELETE":
		req, err = http.NewRequest(httpMethod, url, nil)
		if err != nil {
			return nil, err
		}
		if len(b.queryParams) > 0 {
			req.URL.RawQuery = b.queryParams.Encode()
		}
	}
	req.Header.Set("Accept", "application/json")
	for key, value := range b.headerParams {
		req.Header.Set(key, value)
	}
	return req, nil
}

func (b *GetPhotoDetailsRequestBuilderImpl) Run() (GetPhotoDetailsResponse, error) {
	request, err := b.build()
	if err != nil {
		return nil, err
	}
	request.URL.RawQuery = request.URL.Query().Encode()

	restClient := restclient.GetClient()
	if restClient == nil {
		return nil, fmt.Errorf("A rest client has not been registered yet. You must call client.RegisterClient first")
	}

	if restClient.Debug() {
		restclient.DebugRequest(request)
	}

	response, err := restClient.HttpClient().Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	if restClient.Debug() {
		restclient.DebugResponse(response)
	}

	return NewGetPhotoDetailsResponse(response.Body)
}

func (b *GetPhotoDetailsRequestBuilderImpl) RunAsync(callback GetPhotoDetailsCallback) {
	if callback != nil {
		callback.OnStart()
	}

	go func(b *GetPhotoDetailsRequestBuilderImpl) {
		response, err := b.Run()

		if callback != nil {
			if err != nil {
				callback.OnError(err.Error())
			} else {
				callback.OnSuccess(response)
			}
		}
	}(b)
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "input.go", src, parser.ParseComments)
	assert.NoError(t, err)

	p := parse.NewParser(f, "test")
	result := p.Parse()

	data, err := Generate(result)
	assert.NoError(t, err)

	assert.Equal(t, output, string(data))
}

func TestGetParamsList(t *testing.T) {
	var testCases = []struct {
		input  string
		output string
	}{
		{
			`package main
			func empty() {
			}
			`,
			"",
		},
		{
			`package main
			func oneArgument(arg string) {
			}
			`,
			"arg string",
		},
		{
			`package main
			func secondArgument(arg1 string, arg2 int) {
			}
			`,
			"arg1 string,arg2 int",
		},
		{
			`package main
			func multipleArguments(arg1 string, arg2 int, arg3 bool, arg4 string) {
			}
			`,
			"arg1 string,arg2 int,arg3 bool,arg4 string",
		},
	}

	for _, tc := range testCases {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "input.go", tc.input, 0)
		assert.NoError(t, err)
		paramList := getParamsList(f.Decls[0].(*ast.FuncDecl).Type)
		assert.Equal(t, tc.output, paramList)
	}
}

func TestParamType(t *testing.T) {
	var testCases = []struct {
		input  string
		output string
	}{
		{
			`package main
			func one(a string) {
			}
			`,
			"string",
		},
		{
			`package main
			func two(b *Pointer) {
			}
			`,
			"*Pointer",
		},
		{
			`package main
			func three(b *some.Pointer) {
			}
			`,
			"*some.Pointer",
		},
	}

	for _, tc := range testCases {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, "input.go", tc.input, 0)
		assert.NoError(t, err)
		params := f.Decls[0].(*ast.FuncDecl).Type.Params
		paramType := getParamType(params.List[0].Type)
		assert.Equal(t, tc.output, paramType)
	}
}
