package gphoto

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
	"strconv"
)

const (
	photoBucket   = "photo"
	googlePhotoDB = "gphoto.db"
)

var albumNotExists = errors.New("album not exists")

// BoltRepository is a boltdb repository implementaion.
type BoltRepository struct {
	DB *bbolt.DB
}

func (r BoltRepository) close() error {
	return r.DB.Close()
}

// savePhotos save photos,received via api, into album bucket.
func (r BoltRepository) savePhotos(album string, photos []*GooglePhoto) error {
	tx, err := r.DB.Begin(true)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	photoBucket := tx.Bucket([]byte(photoBucket))

	albumBucket, err := photoBucket.CreateBucketIfNotExists([]byte(album))
	if err != nil {
		return err
	}

	for _, photo := range photos {
		photoID, err := albumBucket.NextSequence()
		if err != nil {
			return err
		}
		if buf, err := json.Marshal(photo); err != nil {
			return err
		} else if err := albumBucket.Put([]byte(strconv.FormatUint(photoID, 10)), buf); err != nil {
			return err
		}
	}
	logrus.WithField("album", album).Debugln("save album photo")
	return tx.Commit()
}

// listPhotos fetch photos from album boltdb bucket.
func (r BoltRepository) listPhotos(album string) ([]*GooglePhoto, error) {
	var (
		items []*GooglePhoto
		photo GooglePhoto
	)
	tx, err := r.DB.Begin(true)
	if err != nil {
		return items, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	pBucket := tx.Bucket([]byte(photoBucket))
	albumBucket := pBucket.Bucket([]byte(album))
	if albumBucket == nil {
		logrus.WithField("album", album).Debugln(albumNotExists)
		return items, albumNotExists
	}

	c := albumBucket.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		if err = json.Unmarshal(v, &photo); err != nil {
			logrus.WithError(err).WithField("album", album).Errorln("unmarshal bolt value error")
			return items, err
		}
		items = append(items, &photo)
	}
	err = tx.Commit()
	logrus.WithField("album", album).Debugln("list album photo from repo")
	return items, err
}

// truncateAlbum truncate boltdb album bucket.
func (r BoltRepository) truncateAlbum(album string) error {
	tx, err := r.DB.Begin(true)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	pBucket := tx.Bucket([]byte(photoBucket))
	albumBucket := pBucket.Bucket([]byte(album))
	if albumBucket == nil {
		return nil
	}
	if err = pBucket.DeleteBucket([]byte(album)); err != nil {
		return err
	}
	logrus.WithField("album", album).Debugln("truncate album photo")
	return tx.Commit()
}

// NewBoltRepository make BoltRepository instance.
func NewBoltRepository(DB *bbolt.DB) *BoltRepository {
	var err error
	err = DB.Update(func(tx *bbolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(photoBucket))
		if err != nil {
			return fmt.Errorf("create photo bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		logrus.Fatalln(err)
	}
	return &BoltRepository{DB: DB}
}
