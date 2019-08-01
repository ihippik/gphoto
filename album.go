package gphoto

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

type GoogleAlbumResponse struct {
	GoogleAlbums []GoogleAlbum `json:"albums"`
}

type GoogleAlbum struct {
	ID                    string `json:"id"`
	Title                 string `json:"title"`
	ProductURL            string `json:"productUrl"`
	MediaItemsCount       string `json:"mediaItemsCount"`
	CoverPhotoBaseURL     string `json:"coverPhotoBaseUrl"`
	CoverPhotoMediaItemID string `json:"coverPhotoMediaItemId"`
}

func (c Client) getAlbumList() ([]GoogleAlbum, error) {
	var (
		googleResponse GoogleAlbumResponse
		albums         []GoogleAlbum
	)

	req, _ := http.NewRequest("GET", getAlbumsURL, nil)
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
			return albums, invalidToken
		default:
			return albums, errors.New(res.Status)
		}
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return albums, err
	}

	if err := json.Unmarshal(body, &googleResponse); err != nil {
		return albums, err
	}
	return googleResponse.GoogleAlbums, nil
}
