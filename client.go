package gphoto

import (
	"errors"
	"github.com/sirupsen/logrus"
)

type Api interface {
	refreshAccessToken(clientID, clientSecret, refreshToken string) (string, error)
	getAlbumList(accessToken string) ([]GoogleAlbum, error)
	searchPhotos(accessToken, albumID string) ([]GooglePhoto, error)
}

type client struct {
	clientID     string
	clientSecret string
	accessToken  string
	refreshToken string
	api          Api
}

// Some api errors.
var (
	refreshTokenErr = errors.New("can`t refresh token")
	searchPhotosErr = errors.New("search photos error")
	getAlbumErr     = errors.New("get album error")
)

// NewGoogleClient create google photo api client.
func NewGoogleClient(clientID, clientSecret, refreshToken string) *client {
	gAPI := newGoogleApi()
	return &client{
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		api:          gAPI,
	}
}

// GetAlbumList fetch all photo albums.
func (c *client) GetAlbumList() ([]GoogleAlbum, error) {
	albums, err := c.api.getAlbumList(c.accessToken)
	if err == unauthorizedErr {
		c.accessToken, err = c.api.refreshAccessToken(c.clientID, c.clientSecret, c.refreshToken)
		if err != nil {
			logrus.WithError(err).Error(refreshTokenErr)
			return albums, refreshTokenErr
		}
		albums, err = c.api.getAlbumList(c.accessToken)
		if err != nil {
			logrus.WithError(err).Errorln(getAlbumErr)
		}
	} else if err != nil {
		logrus.WithError(err).Errorln(getAlbumErr)
	}
	return albums, err
}

// GetPhotoByAlbum fetch photos of a specific album.
func (c *client) GetPhotoByAlbum(albumID string) ([]GooglePhoto, error) {
	photos, err := c.api.searchPhotos(c.accessToken, albumID)
	if err == unauthorizedErr {
		c.accessToken, err = c.api.refreshAccessToken(c.clientID, c.clientSecret, c.refreshToken)
		if err != nil {
			logrus.WithError(err).Error(refreshTokenErr)
			return photos, refreshTokenErr
		}
		photos, err = c.api.searchPhotos(c.accessToken, albumID)
		if err != nil {
			logrus.WithError(err).Errorln(searchPhotosErr)
		}
	} else if err != nil {
		logrus.WithError(err).Errorln(searchPhotosErr)
	}
	return photos, err
}
