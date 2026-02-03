package jesr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"pdf_service_web/models"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Api struct {
	BaseUrl string
}

type GetDocumentsResponse struct {
	Documents []models.Document `json:"documents"`
}

func (t Api) GetDocumentsByOwnerUUID(ownerUUID uuid.UUID, limit, offset uint32) ([]models.Document, error) {
	url := fmt.Sprintf("%s/api/v1/documents?limit=%d&offset=%d&exclude=pdfBase64&ownerUUID=%s", t.BaseUrl, limit, offset, ownerUUID.String())
	method := "GET"

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	response := &GetDocumentsResponse{}
	err = json.Unmarshal(body, response)
	return response.Documents, err
}

func (t Api) DeleteDocuments(documentUUID, ownerUUID uuid.UUID) error {
	url := fmt.Sprintf("%s/api/v1/documents?documentUUID=%s&ownerUUID=%s", t.BaseUrl, documentUUID.String(), ownerUUID.String())
	method := "DELETE"

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

type UploadRequest struct {
	DocumentBase64String string    `json:"documentBase64String"`
	DocumentTitle        string    `json:"documentTitle"`
	OwnerType            int       `json:"ownerType"`
	OwnerUUID            uuid.UUID `json:"ownerUUID"`
}

func (t Api) UploadDocument(request UploadRequest) (uuid.UUID, error) {
	url := fmt.Sprintf("%s/api/v1/documents/", t.BaseUrl)
	method := "POST"

	bytes, err := json.Marshal(request)
	if err != nil {
		return uuid.Nil, err
	}

	body := string(bytes)

	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return uuid.Nil, err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return uuid.Nil, err
	}

	if res.StatusCode != http.StatusOK {
		return uuid.Nil, errors.New("unexpected status code")

	}

	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return uuid.Nil, err
	}

	type ExpectedResponse struct {
		DocumentUUID uuid.UUID `json:"documentUUID"`
	}

	data := &ExpectedResponse{}

	err = json.Unmarshal(resBytes, &data)
	if err != nil {
		return uuid.Nil, err
	}

	return data.DocumentUUID, nil
}

type AddMetaRequest struct {
	DocumentUUID         uuid.UUID `json:"documentUUID"`
	OwnerUUID            uuid.UUID `json:"ownerUUID"`
	OwnerType            int       `json:"ownerType"`
	DocumentBase64String *string   `json:"documentBase64String"`
}

func (t Api) AddMeta(request AddMetaRequest) error {
	url := fmt.Sprintf("%s/api/v1/meta/", t.BaseUrl)
	method := "POST"

	bytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	body := string(bytes)

	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		bytes, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("unexpected status code returned by api: %s", string(bytes))
	}

	return nil
}

var GetMetaNotFoundError = errors.New("user meta data not found")

func (t Api) GetMeta(documentUid, ownerUid string, offset, limit uint32) (models.Meta, error) {
	url := fmt.Sprintf("%s/api/v1/meta/?documentUUID=%s&ownerUUID=%s&offset=%d&limit=%d", t.BaseUrl, documentUid, ownerUid, offset, limit)
	method := "GET"

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return models.Meta{}, err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return models.Meta{}, err
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return models.Meta{}, err
	}

	if res.StatusCode != http.StatusOK {
		switch res.StatusCode {
		case http.StatusNotFound:
			return models.Meta{}, GetMetaNotFoundError

		default:
			return models.Meta{}, fmt.Errorf("unexpected status code returned by api: %s", string(bytes))
		}
	}

	data := &models.Meta{}
	err = json.Unmarshal(bytes, data)
	if err != nil {
		return models.Meta{}, err
	}

	return *data, nil
}

// AddSelectionsBulk Proxies request through to API server. - TODO: Add checks
func (t Api) AddSelectionsBulk(c *gin.Context) (string, error) {
	url := fmt.Sprintf("%s/api/v1/selections/bulk", t.BaseUrl)
	method := "POST"

	req, err := http.NewRequest(method, url, c.Request.Body)
	if err != nil {
		return "", err
	}
	req.Header = c.Request.Header

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code returned by api: %s", string(bytes))
	}

	return string(bytes), nil
}

// GetSelectionListString Proxies request through to API server. - TODO: Add checks
func (t Api) GetSelectionListString(c *gin.Context) (string, error) {
	queryStr := c.Request.URL.RawQuery
	url := fmt.Sprintf("%s/api/v1/selections?%s", t.BaseUrl, queryStr)
	method := "GET"

	req, err := http.NewRequest(method, url, c.Request.Body)
	if err != nil {
		return "", err
	}
	req.Header = c.Request.Header

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code returned by api: %s", string(bytes))
	}

	return string(bytes), nil
}

// DeleteSelection Proxies request through to API server. - TODO: Add checks, add bulk.
func (t Api) DeleteSelection(c *gin.Context) error {
	queryStr := c.Request.URL.RawQuery
	url := fmt.Sprintf("%s/api/v1/selections?%s", t.BaseUrl, queryStr)
	method := "DELETE"

	req, err := http.NewRequest(method, url, c.Request.Body)
	if err != nil {
		return err
	}
	req.Header = c.Request.Header

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code returned by api: %d", res.StatusCode)
	}

	return nil
}
