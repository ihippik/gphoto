package gphoto

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type GooglePhotoResponse struct {
	GooglePhotos []GooglePhoto `json:"photos"`
}

type GooglePhoto struct {
	ID                    string `json:"id"`
	Title                 string `json:"title"`
	ProductURL            string `json:"productUrl"`
	MediaItemsCount       string `json:"mediaItemsCount"`
	CoverPhotoBaseURL     string `json:"coverPhotoBaseUrl"`
	CoverPhotoMediaItemID string `json:"coverPhotoMediaItemId"`
}

func (c Client) searchPhotos(albumID string) ([]GooglePhoto, error) {
	var (
		googleResponse GooglePhotoResponse
		photos         []GooglePhoto
	)

	form := url.Values{}
	form.Add("pageSize", "100")
	form.Add("albumId", albumID)

	req, err := http.NewRequest("POST", searchPhotoURL, strings.NewReader(form.Encode()))
	if err != nil {
		return photos, err
	}

	auth := fmt.Sprintf("Bearer %s", c.accessToken)
	req.Header.Add("Authorization", auth)
	req.Header.Add("cache-control", "no-cache")

	res, _ := http.DefaultClient.Do(req)
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		logrus.Errorln(res.Status)
		switch res.StatusCode {
		case 401, 403:
			return photos, invalidToken
		default:
			return photos, errors.New(res.Status)
		}
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return photos, err
	}

	if err := json.Unmarshal(body, &googleResponse); err != nil {
		return photos, err
	}
	return googleResponse.GooglePhotos, err
}
