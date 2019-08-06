package gphoto

import "time"

type googlePhotoResponse struct {
	GooglePhotos []*GooglePhoto `json:"mediaItems"`
}

// GooglePhoto represent google album structure received from api.
type GooglePhoto struct {
	ID            string `json:"id"`
	ProductURL    string `json:"productUrl"`
	BaseURL       string `json:"baseUrl"`
	MimeType      string `json:"mimeType"`
	MediaMetadata struct {
		CreationTime time.Time `json:"creationTime"`
		Width        string    `json:"width"`
		Height       string    `json:"height"`
		Photo        struct {
			CameraMake      string  `json:"cameraMake"`
			CameraModel     string  `json:"cameraModel"`
			FocalLength     int     `json:"focalLength"`
			ApertureFNumber float64 `json:"apertureFNumber"`
			IsoEquivalent   int     `json:"isoEquivalent"`
		} `json:"photo"`
	} `json:"mediaMetadata"`
	Filename string `json:"filename"`
}
