package music

type DeezerResponse struct {
	Data []DeezerResult `json:"data"`
}

type DeezerResult struct {
	ID                    int          `json:"id"`
	Readable              bool         `json:"readable"`
	Title                 string       `json:"title"`
	TitleShort            string       `json:"title_short"`
	Link                  string       `json:"link"`
	Duration              int          `json:"duration"`
	Rank                  int          `json:"rank"`
	ExplicitLyrics        bool         `json:"explicit_lyrics"`
	ExplicitContentLyrics int          `json:"explicit_content_lyrics"`
	ExplicitContentCover  int          `json:"explicit_content_cover"`
	Preview               string       `json:"preview"`
	Md5Image              string       `json:"md5_image"`
	Artist                DeezerArtist `json:"artist"`
	Album                 DeezerAlbum  `json:"album"`
	Type                  string       `json:"type"`
}

type DeezerArtist struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Link          string `json:"link"`
	Picture       string `json:"picture"`
	PictureSmall  string `json:"picture_small"`
	PictureMedium string `json:"picture_medium"`
	PictureBig    string `json:"picture_big"`
	PictureXl     string `json:"picture_xl"`
	Tracklist     string `json:"tracklist"`
	Type          string `json:"type"`
}

type DeezerAlbum struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Cover       string `json:"cover"`
	CoverSmall  string `json:"cover_small"`
	CoverMedium string `json:"cover_medium"`
	CoverBig    string `json:"cover_big"`
	CoverXl     string `json:"cover_xl"`
	Md5Image    string `json:"md5_image"`
	Tracklist   string `json:"tracklist"`
	Type        string `json:"type"`
}
