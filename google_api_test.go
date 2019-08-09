package gphoto

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGoogleApi(t *testing.T) {
	want := &googleApi{
		getAlbumsURL:   "https://photoslibrary.googleapis.com/v1/albums",
		searchPhotoURL: "https://photoslibrary.googleapis.com/v1/mediaItems:search",
		getTokenURL:    "https://accounts.google.com/o/oauth2/token",
	}
	if got := NewGoogleApi(); !reflect.DeepEqual(got, want) {
		t.Errorf("NewGoogleApi() = %v, want %v", got, want)
	}
}

func Test_urlIsValid(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		url        string
		statusCode int
		want       bool
	}{
		{
			name:       "is valid",
			url:        "/check",
			statusCode: http.StatusOK,
			want:       true,
		},
		{
			name:       "invalid",
			url:        "/check",
			statusCode: http.StatusUnauthorized,
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(tt.statusCode)
				assert.Equal(t, req.URL.String(), tt.url)
			}))
			defer server.Close()
			api := googleApi{client: server.Client(), searchPhotoURL: server.URL + tt.path}
			if got := api.urlIsValid(server.URL + tt.url); got != tt.want {
				t.Errorf("urlIsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_googleApi_searchPhotos(t *testing.T) {
	type fields struct {
		getAlbumsURL string
		statusCode   int
		payload      []byte
	}
	type args struct {
		accessToken string
		albumID     string
		path        string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr error
	}{
		{
			name: "StatusOK",
			fields: fields{
				statusCode: http.StatusOK,
				payload:    []byte(`{"mediaItems":[{"id":"photoid","baseUrl":"photo_path.png"}]}`),
			},
			args: args{
				accessToken: "accesstoken",
				path:        "/search-photo",
			},
			want:    "photo_path.png",
			wantErr: nil,
		},
		{
			name: "StatusUnauthorized",
			fields: fields{
				statusCode: http.StatusUnauthorized,
			},
			args: args{
				accessToken: "accesstoken",
				path:        "/search-photo",
			},
			wantErr: unauthorizedErr,
		},
		{
			name: "StatusNotFound",
			fields: fields{
				statusCode: http.StatusNotFound,
			},
			args: args{
				accessToken: "accesstoken",
				path:        "/get-album-list",
			},
			wantErr: errors.New("404 Not Found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				// Test request parameters
				rw.WriteHeader(tt.fields.statusCode)
				assert.Equal(t, req.URL.String(), tt.args.path)
				_, _ = rw.Write(tt.fields.payload)
			}))
			defer server.Close()

			api := googleApi{client: server.Client(), searchPhotoURL: server.URL + tt.args.path}
			photos, err := api.searchPhotos(tt.args.accessToken, tt.args.albumID)
			if err == nil {
				assert.Len(t, photos, 1)
				assert.Equal(t, photos[0].BaseURL, tt.want)
			} else {
				assert.Equal(t, err, tt.wantErr)
			}
		})
	}
}

func Test_googleApi_getAlbumList(t *testing.T) {
	type fields struct {
		getAlbumsURL string
		statusCode   int
		payload      []byte
	}
	type args struct {
		accessToken string
		path        string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr error
	}{
		{
			name: "StatusOK",
			fields: fields{
				statusCode: http.StatusOK,
				payload:    []byte(`{"albums":[{"id":"albumid","title":"album title"}]}`),
			},
			args: args{
				accessToken: "accesstoken",
				path:        "/get-album-list",
			},
			want:    "album title",
			wantErr: nil,
		},
		{
			name: "StatusUnauthorized",
			fields: fields{
				statusCode: http.StatusUnauthorized,
			},
			args: args{
				accessToken: "accesstoken",
				path:        "/get-album-list",
			},
			wantErr: unauthorizedErr,
		},
		{
			name: "StatusNotFound",
			fields: fields{
				statusCode: http.StatusNotFound,
			},
			args: args{
				accessToken: "accesstoken",
				path:        "/get-album-list",
			},
			wantErr: errors.New("404 Not Found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				// Test request parameters
				rw.WriteHeader(tt.fields.statusCode)
				assert.Equal(t, req.URL.String(), tt.args.path)
				_, _ = rw.Write(tt.fields.payload)
			}))
			defer server.Close()

			api := googleApi{client: server.Client(), getAlbumsURL: server.URL + tt.args.path}
			albums, err := api.getAlbumList(tt.args.accessToken)
			if err == nil {
				assert.Len(t, albums, 1)
				assert.Equal(t, albums[0].Title, tt.want)
			} else {
				assert.Equal(t, err, tt.wantErr)
			}
		})
	}
}

func Test_googleApi_refreshAccessToken(t *testing.T) {
	type fields struct {
		getTokenURL string
		statusCode  int
		payload     []byte
	}
	type args struct {
		clientID     string
		clientSecret string
		refreshToken string
		path         string
		wantErr      error
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr error
	}{
		{
			name: "statusOK",
			fields: fields{
				statusCode: http.StatusOK,
				payload:    []byte(`{"access_token":"mytoken"}`),
			},
			args: args{
				clientID:     "CLIENT_ID",
				clientSecret: "SECRET_ID",
				refreshToken: "TOKEN",
				path:         "/get-token",
				wantErr:      nil,
			},
			want:    "mytoken",
			wantErr: nil,
		},
		{
			name: "StatusInternalServerError",
			fields: fields{
				statusCode: http.StatusInternalServerError,
			},
			args: args{
				clientID:     "CLIENT_ID",
				clientSecret: "SECRET_ID",
				refreshToken: "TOKEN",
				path:         "/get-token",
			},
			want:    "",
			wantErr: badStatusErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(tt.fields.statusCode)
					assert.Equal(t, req.URL.String(), tt.args.path)
					_, _ = rw.Write(tt.fields.payload)
				}))
				defer server.Close()

				api := googleApi{client: server.Client(), getTokenURL: server.URL + tt.args.path}
				token, err := api.refreshAccessToken(tt.args.clientID, tt.args.clientSecret, tt.args.refreshToken)
				if err == nil {
					assert.Equal(t, tt.want, token)
				} else {
					assert.Equal(t, err, tt.wantErr)
				}
			})
		})
	}
}
