package gphoto

import "errors"

type Client struct {
	accessToken  string
	refreshToken string
}

const getAlbumsURL = "https://photoslibrary.googleapis.com/v1/albums"
const searchPhotoURL = "https://photoslibrary.googleapis.com/v1/mediaItems:search"

var invalidToken = errors.New("invalid token")

func NewClient(accessToken string, refreshToken string) *Client {
	return &Client{accessToken: accessToken, refreshToken: refreshToken}
}
