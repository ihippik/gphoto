package gphoto

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type (
	MockedApi struct {
		mock.Mock
	}
	MockedRepo struct {
		mock.Mock
	}
)

var (
	someErr  = errors.New("some error")
	apiMock  = new(MockedApi)
	repoMock = new(MockedRepo)

	setupSavePhotos = func(err error) {
		repoMock.On("savePhotos", mock.Anything, mock.Anything).Return(err).Once()
	}
	setupTruncateAlbum = func(err error) {
		repoMock.On("truncateAlbum", mock.Anything).Return(err).Once()
	}
	setupListPhotos = func(list []*GooglePhoto, err error) {
		repoMock.On("listPhotos", mock.Anything).Return(list, err).Once()
	}

	setupRefreshAccessToken = func(token string, err error) {
		apiMock.On("refreshAccessToken", mock.Anything, mock.Anything, mock.Anything).Return(token, err).Once()
	}
	setupGetAlbumList = func(list []*GoogleAlbum, err error) {
		apiMock.On("getAlbumList", mock.Anything).Return(list, err).Once()
	}
	setupSearchPhotos = func(list []*GooglePhoto, err error) {
		apiMock.On("searchPhotos", mock.Anything, mock.Anything).Return(list, err).Once()
	}
	setupUrlIsValid = func(result bool) {
		apiMock.On("urlIsValid", mock.Anything).Return(result).Once()
	}
)

func (m MockedRepo) savePhotos(album string, photo []*GooglePhoto) error {
	args := m.Called(album, photo)
	return args.Error(0)
}

func (m MockedRepo) listPhotos(album string) ([]*GooglePhoto, error) {
	args := m.Called(album)
	return args.Get(0).([]*GooglePhoto), args.Error(1)
}

func (m MockedRepo) truncateAlbum(album string) error {
	args := m.Called(album)
	return args.Error(0)

}

func (m MockedRepo) close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockedApi) refreshAccessToken(clientID, clientSecret, refreshToken string) (string, error) {
	args := m.Called(clientID, clientSecret, refreshToken)
	return args.String(0), args.Error(1)
}

func (m *MockedApi) getAlbumList(accessToken string) ([]*GoogleAlbum, error) {
	args := m.Called(accessToken)
	return args.Get(0).([]*GoogleAlbum), args.Error(1)
}

func (m *MockedApi) searchPhotos(accessToken, albumID string) ([]*GooglePhoto, error) {
	args := m.Called(accessToken)
	return args.Get(0).([]*GooglePhoto), args.Error(1)
}

func (m *MockedApi) urlIsValid(url string) bool {
	args := m.Called(url)
	return args.Bool(0)
}

func TestClient_GetAlbumList(t *testing.T) {
	type fields struct {
		clientID     string
		clientSecret string
		accessToken  string
		refreshToken string
	}

	list := []*GoogleAlbum{
		{
			ID:                    "abcdef",
			Title:                 "title",
			ProductURL:            "http://photo.com",
			MediaItemsCount:       "http://photo.com",
			CoverPhotoBaseURL:     "http://photo.com",
			CoverPhotoMediaItemID: "abcdef",
		},
	}

	tests := []struct {
		name    string
		setup   func()
		fields  fields
		want    []*GoogleAlbum
		wantErr error
	}{
		{
			name: "success",
			fields: fields{
				clientID:     "CLIENT_ID",
				clientSecret: "SECRET_ID",
				refreshToken: "TOKEN",
				accessToken:  "ACCESS_TOKEN",
			},
			want:    list,
			wantErr: nil,
			setup: func() {
				setupGetAlbumList(list, nil)
			},
		},
		{
			name: "unauthorizedErr & success",
			fields: fields{
				clientID:     "CLIENT_ID",
				clientSecret: "SECRET_ID",
				refreshToken: "TOKEN",
				accessToken:  "ACCESS_TOKEN",
			},
			want:    list,
			wantErr: nil,
			setup: func() {
				setupGetAlbumList(list, unauthorizedErr)
				setupRefreshAccessToken("token", nil)
				setupGetAlbumList(list, nil)
			},
		},
		{
			name: "unauthorizedErr & fail",
			fields: fields{
				clientID:     "CLIENT_ID",
				clientSecret: "SECRET_ID",
				refreshToken: "TOKEN",
				accessToken:  "ACCESS_TOKEN",
			},
			want:    list,
			wantErr: refreshTokenErr,
			setup: func() {
				setupGetAlbumList(list, unauthorizedErr)
				setupRefreshAccessToken("token", badStatusErr)
			},
		},
		{
			name: "unauthorizedErr & second get albums fail",
			fields: fields{
				clientID:     "CLIENT_ID",
				clientSecret: "SECRET_ID",
				refreshToken: "TOKEN",
				accessToken:  "ACCESS_TOKEN",
			},
			want:    list,
			wantErr: getAlbumErr,
			setup: func() {
				setupGetAlbumList(list, unauthorizedErr)
				setupRefreshAccessToken("token", nil)
				setupGetAlbumList(list, unauthorizedErr)
			},
		},
		{
			name: "get albums fail",
			fields: fields{
				clientID:     "CLIENT_ID",
				clientSecret: "SECRET_ID",
				refreshToken: "TOKEN",
				accessToken:  "ACCESS_TOKEN",
			},
			want:    list,
			wantErr: getAlbumErr,
			setup: func() {
				setupGetAlbumList(list, someErr)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer apiMock.AssertExpectations(t)
			c := &Client{
				clientID:     tt.fields.clientID,
				clientSecret: tt.fields.clientSecret,
				accessToken:  tt.fields.accessToken,
				refreshToken: tt.fields.refreshToken,
				api:          apiMock,
				repo:         repoMock,
			}
			got, err := c.GetAlbumList()
			if err == nil {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetAlbumList() got = %v, want %v", got, tt.want)
				}
			} else {
				assert.Equal(t, tt.wantErr, err)
			}

		})
	}
}

func TestClient_GetPhotoByAlbum(t *testing.T) {
	type fields struct {
		clientID     string
		clientSecret string
		accessToken  string
		refreshToken string
		api          api
		repo         repository
	}

	type args struct {
		albumID string
	}

	list := []*GooglePhoto{
		{
			ID:         "abcdef",
			ProductURL: "http://photo.com",
		},
	}

	repoMock.On("close").Return(nil)

	tests := []struct {
		name    string
		setup   func()
		fields  fields
		args    args
		want    []*GooglePhoto
		wantErr error
	}{
		{
			name: "get photos success",
			fields: fields{
				clientID:     "CLIENT_ID",
				clientSecret: "SECRET_ID",
				refreshToken: "TOKEN",
				accessToken:  "ACCESS_TOKEN",
			},
			args: args{albumID: "asdef"},
			want: list,
			setup: func() {
				setupListPhotos(list, nil)
				setupUrlIsValid(true)
			},
		},
		{
			name: "get photos from api",
			fields: fields{
				clientID:     "CLIENT_ID",
				clientSecret: "SECRET_ID",
				refreshToken: "TOKEN",
				accessToken:  "ACCESS_TOKEN",
			},
			args: args{albumID: "asdef"},
			want: list,
			setup: func() {
				setupListPhotos(list, someErr)
				setupUrlIsValid(false)
				setupSearchPhotos(list, nil)
				setupTruncateAlbum(nil)
				setupSavePhotos(nil)
			},
		},
		{
			name: "get photos from api, unauthorizedErr, success",
			fields: fields{
				clientID:     "CLIENT_ID",
				clientSecret: "SECRET_ID",
				refreshToken: "TOKEN",
				accessToken:  "ACCESS_TOKEN",
			},
			args: args{albumID: "asdef"},
			want: list,
			setup: func() {
				setupListPhotos(list, someErr)
				setupUrlIsValid(false)
				setupSearchPhotos(list, unauthorizedErr)
				setupRefreshAccessToken("token", nil)
				setupSearchPhotos(list, nil)
				setupTruncateAlbum(nil)
				setupSavePhotos(nil)
			},
		},
		{
			name: "get photos from api, unauthorizedErr, search error",
			fields: fields{
				clientID:     "CLIENT_ID",
				clientSecret: "SECRET_ID",
				refreshToken: "TOKEN",
				accessToken:  "ACCESS_TOKEN",
			},
			args:    args{albumID: "asdef"},
			wantErr: searchPhotosErr,
			want:    list,
			setup: func() {
				setupListPhotos(list, someErr)
				setupUrlIsValid(false)
				setupSearchPhotos(list, unauthorizedErr)
				setupRefreshAccessToken("token", nil)
				setupSearchPhotos(list, unauthorizedErr)
			},
		},
		{
			name: "get photos from api, unauthorizedErr, fail",
			fields: fields{
				clientID:     "CLIENT_ID",
				clientSecret: "SECRET_ID",
				refreshToken: "TOKEN",
				accessToken:  "ACCESS_TOKEN",
			},
			args:    args{albumID: "asdef"},
			want:    list,
			wantErr: refreshTokenErr,
			setup: func() {
				setupListPhotos(list, someErr)
				setupUrlIsValid(false)
				setupSearchPhotos(list, unauthorizedErr)
				setupRefreshAccessToken("token", badStatusErr)
			},
		},
		{
			name: "get photos from api, fail",
			fields: fields{
				clientID:     "CLIENT_ID",
				clientSecret: "SECRET_ID",
				refreshToken: "TOKEN",
				accessToken:  "ACCESS_TOKEN",
			},
			args:    args{albumID: "asdef"},
			want:    list,
			wantErr: searchPhotosErr,
			setup: func() {
				setupListPhotos(list, someErr)
				setupUrlIsValid(false)
				setupSearchPhotos(list, someErr)
			},
		},
		{
			name: "truncate album error",
			fields: fields{
				clientID:     "CLIENT_ID",
				clientSecret: "SECRET_ID",
				refreshToken: "TOKEN",
				accessToken:  "ACCESS_TOKEN",
			},
			args:    args{albumID: "asdef"},
			want:    list,
			wantErr: truncateErr,
			setup: func() {
				setupListPhotos(list, nil)
				setupUrlIsValid(false)
				setupSearchPhotos(list, nil)
				setupTruncateAlbum(someErr)
			},
		},
		{
			name: "save photo error",
			fields: fields{
				clientID:     "CLIENT_ID",
				clientSecret: "SECRET_ID",
				refreshToken: "TOKEN",
				accessToken:  "ACCESS_TOKEN",
			},
			args:    args{albumID: "asdef"},
			want:    list,
			wantErr: saveErr,
			setup: func() {
				setupListPhotos(list, nil)
				setupUrlIsValid(false)
				setupSearchPhotos(list, nil)
				setupTruncateAlbum(nil)
				setupSavePhotos(someErr)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer apiMock.AssertExpectations(t)
			defer repoMock.AssertExpectations(t)
			c := &Client{
				clientID:     tt.fields.clientID,
				clientSecret: tt.fields.clientSecret,
				accessToken:  tt.fields.accessToken,
				refreshToken: tt.fields.refreshToken,
				api:          apiMock,
				repo:         repoMock,
			}
			got, err := c.GetPhotoByAlbum(tt.args.albumID)
			if err == nil {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("GetPhotoByAlbum() got = %v, want %v", got, tt.want)
				}
			} else {
				assert.Equal(t, tt.wantErr, err)
			}
		})
	}
}
