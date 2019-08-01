# Google Photo client

A simple client for receiving photos and albums via API Google Photos.

### Example

    import 	"github.com/ihippik/gphoto"
    
    refreshToken := "REFRESH_TOKEN
    clientID := "CLIENT_ID"
    clientSecret := "CLIENT_SECRET"
    client := gphoto.NewGoogleClient(clientID, clientSecret, refreshToken)
    
    albums, err := client.GetAlbumList()
    photos, err:= client.GetPhotoByAlbum(albumID)