package gphoto

type googleAlbumResponse struct {
	GoogleAlbums []*GoogleAlbum `json:"albums"`
}

// GoogleAlbum represent google album structure received from api.
type GoogleAlbum struct {
	ID                    string `json:"id"`
	Title                 string `json:"title"`
	ProductURL            string `json:"productUrl"`
	MediaItemsCount       string `json:"mediaItemsCount"`
	CoverPhotoBaseURL     string `json:"coverPhotoBaseUrl"`
	CoverPhotoMediaItemID string `json:"coverPhotoMediaItemId"`
}
