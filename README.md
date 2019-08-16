# Google Photo client

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Build Status](https://travis-ci.com/ihippik/gphoto.svg?branch=master)](https://travis-ci.com/ihippik/gphoto)
[![codecov](https://codecov.io/gh/ihippik/gphoto/branch/master/graph/badge.svg)](https://codecov.io/gh/ihippik/gphoto)

A simple client for receiving photos and albums via API Google Photos.

Photos received from Google Api are cached in the database (bolt db by default), for the duration of the link itself (one hour).

### Example
```go
import 	"github.com/ihippik/gphoto"

refreshToken := "REFRESH_TOKEN"
clientID := "CLIENT_ID"
clientSecret := "CLIENT_SECRET"
client,err := gphoto.NewGoogleClient(clientID, clientSecret, refreshToken)

albums, err := client.GetAlbumList()
photos, err:= client.GetPhotoByAlbum(albumID)
```
