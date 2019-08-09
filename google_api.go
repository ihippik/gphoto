package gphoto

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type googleApi struct {
	client         *http.Client
	getAlbumsURL   string
	searchPhotoURL string
	getTokenURL    string
}

type refreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

const defaultLimit = 100

var (
	unauthorizedErr = errors.New("unauthorized")
	badStatusErr    = errors.New("bad status")
)

// NewGoogleApi represent client for low-level requests to Google Photo Api.
func NewGoogleApi() *googleApi {
	return &googleApi{
		client:         http.DefaultClient,
		getAlbumsURL:   "https://photoslibrary.googleapis.com/v1/albums",
		searchPhotoURL: "https://photoslibrary.googleapis.com/v1/mediaItems:search",
		getTokenURL:    "https://accounts.google.com/o/oauth2/token",
	}
}

func (g *googleApi) refreshAccessToken(clientID, clientSecret, refreshToken string) (string, error) {
	const refreshTokenType = "refresh_token"
	var refreshResponse refreshResponse

	data := url.Values{}
	data.Set("grant_type", refreshTokenType)
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	req, err := http.NewRequest("POST", g.getTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return refreshResponse.AccessToken, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("cache-control", "no-cache")
	res, err := g.client.Do(req)
	if err != nil {
		return refreshResponse.AccessToken, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		logrus.WithField("status", res.StatusCode).Errorln("bad status")
		return refreshResponse.AccessToken, badStatusErr
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return refreshResponse.AccessToken, err
	}
	if err := json.Unmarshal(body, &refreshResponse); err != nil {
		return refreshResponse.AccessToken, err
	}
	logrus.WithFields(logrus.Fields{
		"expires": refreshResponse.ExpiresIn,
	}).Infoln("access token refreshed")

	return refreshResponse.AccessToken, nil
}

func (g *googleApi) getAlbumList(accessToken string) ([]*GoogleAlbum, error) {
	var (
		googleResponse googleAlbumResponse
		albums         []*GoogleAlbum
	)

	req, err := http.NewRequest("GET", g.getAlbumsURL, nil)
	if err != nil {
		return albums, err
	}
	auth := fmt.Sprintf("Bearer %s", accessToken)
	req.Header.Add("Authorization", auth)
	req.Header.Add("cache-control", "no-cache")

	res, err := g.client.Do(req)
	if err != nil {
		return albums, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return albums, unauthorizedErr
	default:
		return albums, errors.New(res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return albums, err
	}

	if err = json.Unmarshal(body, &googleResponse); err != nil {
		return albums, err
	}
	return googleResponse.GoogleAlbums, nil
}

func (g *googleApi) searchPhotos(accessToken, albumID string) ([]*GooglePhoto, error) {
	var (
		googleResponse googlePhotoResponse
		photos         []*GooglePhoto
	)

	form := url.Values{}
	form.Set("pageSize", strconv.Itoa(defaultLimit))
	form.Set("albumId", albumID)
	req, err := http.NewRequest("POST", g.searchPhotoURL, strings.NewReader(form.Encode()))
	if err != nil {
		return photos, err
	}

	auth := fmt.Sprintf("Bearer %s", accessToken)
	req.Header.Add("Authorization", auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("cache-control", "no-cache")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return photos, err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		return photos, unauthorizedErr
	default:
		return photos, errors.New(res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return photos, err
	}

	if err := json.Unmarshal(body, &googleResponse); err != nil {
		return photos, err
	}

	logrus.WithFields(logrus.Fields{
		"album": albumID,
		"count": len(googleResponse.GooglePhotos),
	}).Debugln("get album photo from api")
	return googleResponse.GooglePhotos, err
}

// urlIsValid check the link to the photo still expired.
func (g *googleApi) urlIsValid(url string) bool {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}
	res, err := g.client.Do(req)
	if err != nil {
		return false
	}
	defer func() {
		_ = res.Body.Close()
	}()

	return res.StatusCode == http.StatusOK
}
