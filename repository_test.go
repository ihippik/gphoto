package gphoto

import (
	"os"
	"reflect"
	"testing"

	"encoding/json"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
)

var db *bbolt.DB

func TestBoltRepository_savePhotos(t *testing.T) {

	type args struct {
		album  string
		photos []*GooglePhoto
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				album: "test",
				photos: []*GooglePhoto{
					{
						BaseURL: "http://test.ts",
					},
				},
			},
			wantErr: false,
		},
	}

	Setup(t)
	r := BoltRepository{
		DB: db,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := r.savePhotos(tt.args.album, tt.args.photos); (err != nil) != tt.wantErr {
				t.Errorf("savePhotos() error = %v, wantErr %v", err, tt.wantErr)
			}
			db.View(func(tx *bbolt.Tx) error {
				// Assume bucket exists and has keys
				b := tx.Bucket([]byte(photoBucket))
				albumBucket := b.Bucket([]byte(tt.args.album))
				if albumBucket == nil {
					t.Error(t, "album bucket is empty")
				}
				c := albumBucket.Cursor()
				var photos []*GooglePhoto
				for k, v := c.First(); k != nil; k, v = c.Next() {
					var photo GooglePhoto
					err := json.Unmarshal(v, &photo)
					if err != nil {
						assert.Error(t, err)
					}
					photos = append(photos, &photo)
				}
				if !reflect.DeepEqual(tt.args.photos, photos) {
					t.Error("result not equal")
				}
				return nil
			})
		})
	}
}

func TestBoltRepository_truncateAlbum(t *testing.T) {
	type args struct {
		album string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				album: "temp",
			},
			wantErr: false,
		},
	}

	Setup(t)
	r := BoltRepository{
		DB: db,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.Update(func(tx *bbolt.Tx) error {
				pb := tx.Bucket([]byte(photoBucket))
				albumBucket, err := pb.CreateBucketIfNotExists([]byte(tt.args.album))
				if err != nil {
					return err
				}
				err = albumBucket.Put([]byte("photo"), []byte("png"))
				return err
			})
			if err != nil {
				assert.Error(t, err)
			}
			if err := r.truncateAlbum(tt.args.album); (err != nil) != tt.wantErr {
				t.Errorf("truncateAlbum() error = %v, wantErr %v", err, tt.wantErr)
			}
			err = db.View(func(tx *bbolt.Tx) error {
				b := tx.Bucket([]byte(photoBucket))
				album := b.Bucket([]byte(tt.args.album))
				if album != nil {
					t.Error("album not empty")
				}
				return nil
			})
			if err != nil {
				assert.Error(t, err)
			}
		})
	}
}

func TestBoltRepository_listPhotos(t *testing.T) {
	type args struct {
		album string
	}
	tests := []struct {
		name    string
		args    args
		want    []*GooglePhoto
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				album: "temp_list",
			},
			want:    nil,
			wantErr: false,
		},
	}
	Setup(t)
	r := BoltRepository{
		DB: db,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.Update(func(tx *bbolt.Tx) error {
				pb := tx.Bucket([]byte(photoBucket))
				albumBucket, err := pb.CreateBucketIfNotExists([]byte(tt.args.album))
				if err != nil {
					return err
				}
				for _, p := range tt.want {
					data, err := json.Marshal(p)
					if err != nil {
						t.Error(err)
					}
					err = albumBucket.Put([]byte("1"), data)
				}
				return err
			})
			got, err := r.listPhotos(tt.args.album)
			if (err != nil) != tt.wantErr {
				t.Errorf("listPhotos() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listPhotos() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Setup(t *testing.T) {
	t.Helper()
	var err error

	db, err = initDB("test.db")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucket([]byte(photoBucket))
		return err
	})
	if err != nil {
		t.Error("can`t create photoBucket")
	}

	defer func() {
		err := os.Remove("test.db")
		if err != nil {
			assert.FailNow(t, err.Error())
		}
	}()
}

func TestNewBoltRepository(t *testing.T) {

	Setup(t)
	tests := []struct {
		name    string
		want    *BoltRepository
		wantErr bool
	}{
		{
			name: "success",
			want: &BoltRepository{
				DB: db,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewBoltRepository(db)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBoltRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBoltRepository() got = %v, want %v", got, tt.want)
			}
			db.View(func(tx *bbolt.Tx) error {
				// Assume bucket exists and has keys
				b := tx.Bucket([]byte(photoBucket))
				assert.NotNil(t, b)
				return nil
			})
		})
	}
}
