package gphoto

import (
	"errors"

	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

type api interface {
	refreshAccessToken(clientID, clientSecret, refreshToken string) (string, error)
	getAlbumList(accessToken string) ([]*GoogleAlbum, error)
	searchPhotos(accessToken, albumID string) ([]*GooglePhoto, error)
	urlIsValid(url string) bool
}

type repository interface {
	savePhotos(album string, photo []*GooglePhoto) error
	listPhotos(album string) ([]*GooglePhoto, error)
	truncateAlbum(album string) error
	close() error
}

// Client struct.
type Client struct {
	clientID     string
	clientSecret string
	accessToken  string
	refreshToken string
	api          api
	repo         repository
}

// Some api errors.
var (
	refreshTokenErr = errors.New("can`t refresh token")
	searchPhotosErr = errors.New("search photos error")
	getAlbumErr     = errors.New("get album error")
	truncateErr     = errors.New("truncate album error")
	saveErr         = errors.New("save album error")
)

// NewGoogleClient create google photo api Client.
func NewGoogleClient(clientID, clientSecret, refreshToken string) *Client {
	db, err := initDB(googlePhotoDB)
	if err != nil {
		logrus.WithError(err).Fatalln("boltDB init error")
	}

	repo := NewBoltRepository(db)
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		api:          NewGoogleApi(),
		repo:         repo,
	}
}

// GetAlbumList fetch all photo albums.
func (c *Client) GetAlbumList() ([]*GoogleAlbum, error) {
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
			return albums, getAlbumErr
		}
	} else if err != nil {
		logrus.WithError(err).Errorln(getAlbumErr)
		return albums, getAlbumErr
	}
	return albums, err
}

// GetPhotoByAlbum fetch photos of a specific album.
func (c *Client) GetPhotoByAlbum(albumID string) ([]*GooglePhoto, error) {
	var (
		photos []*GooglePhoto
		err    error
	)

	photos, err = c.repo.listPhotos(albumID)
	urlIsValid := len(photos) > 0 && c.api.urlIsValid(photos[0].BaseURL)
	logrus.WithField("isValid", urlIsValid).Debugln("first photo url is valid")
	if err == nil && urlIsValid {
		return photos, nil
	} else {
		photos, err = c.api.searchPhotos(c.accessToken, albumID)
		if err == unauthorizedErr {
			c.accessToken, err = c.api.refreshAccessToken(c.clientID, c.clientSecret, c.refreshToken)
			if err != nil {
				logrus.WithError(err).Error(refreshTokenErr)
				return photos, refreshTokenErr
			}
			photos, err = c.api.searchPhotos(c.accessToken, albumID)
			if err != nil {
				logrus.WithError(err).Errorln(searchPhotosErr)
				return photos, searchPhotosErr
			}
		} else if err != nil {
			logrus.WithError(err).Errorln(searchPhotosErr)
			return photos, searchPhotosErr
		}

		if err = c.repo.truncateAlbum(albumID); err != nil {
			logrus.WithError(err).Error(truncateErr)
			return photos, truncateErr
		}

		if len(photos) > 0 {
			if err = c.repo.savePhotos(albumID, photos); err != nil {
				logrus.WithError(err).Errorln(saveErr)
				return photos, saveErr
			}
		}
	}

	return photos, err
}

// Close DB repository connection.
func (c *Client) Close() error {
	return c.repo.close()
}

// initDB init Bolt database connection.
func initDB(dbName string) (*bolt.DB, error) {
	logrus.Debugln("bolt db connection open")
	return bolt.Open(dbName, 0600, nil)
}
