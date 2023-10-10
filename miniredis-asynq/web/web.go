package web

import (
	"embed"
	_ "embed"
)

//go:embed static/*
var EmbedDirStatic embed.FS

//go:embed pages/*.gohtml layouts/*.gohtml components/*.gohtml
var EmbedDirViews embed.FS
