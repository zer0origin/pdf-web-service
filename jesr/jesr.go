package jesr

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strings"
)

type Api struct {
	BaseUrl string
}

type GetDocumentsResponse struct {
	Documents []Document `json:"documents"`
}

func (t Api) GetDocuments(uid uuid.UUID) ([]Document, error) {
	url := fmt.Sprintf("%s/api/v1/documents?exclude=pdfBase64&ownerUUID=%s", t.BaseUrl, uid.String())
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

type UploadRequest struct {
	DocumentBase64String string    `json:"documentBase64String"`
	DocumentTitle        string    `json:"documentTitle"`
	OwnerType            int       `json:"ownerType"`
	OwnerUUID            uuid.UUID `json:"ownerUUID"`
}

func (t Api) UploadDocument(request UploadRequest) error {
	url := fmt.Sprintf("%s/api/v1/documents/", t.BaseUrl)
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
		return errors.New("unexpected status code")
	}

	return nil
}
