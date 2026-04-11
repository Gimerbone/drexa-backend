package app

import "os"

var Host = os.Getenv("HOST")
var Port = os.Getenv("PORT")
