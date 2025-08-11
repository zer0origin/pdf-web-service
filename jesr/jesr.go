package jesr

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
)

type Api struct {
	BaseUrl string
}

type GetDocumentsResponse struct {
	Documents []Document `json:"documents"`
}

func (t Api) GetDocuments(uid uuid.UUID) ([]Document, error) {
	url := fmt.Sprintf("%s/api/v1/documents?ownerUUID=%s", t.BaseUrl, uid.String())
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
	fmt.Println(string(body))

	response := &GetDocumentsResponse{}
	err = json.Unmarshal(body, response)
	return response.Documents, err
}
