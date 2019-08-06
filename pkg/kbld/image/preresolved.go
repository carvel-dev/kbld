package image

type PreresolvedImage struct {
	url string
}

type PreresolvedImageSourceURL struct {
	Type string // always set to 'preresolved'
	URL  string
}

func (PreresolvedImageSourceURL) meta() {}

func NewPreresolvedImage(url string) PreresolvedImage {
	return PreresolvedImage{url}
}

func (i PreresolvedImage) URL() (string, []ImageMeta, error) {
	return i.url, []ImageMeta{PreresolvedImageSourceURL{Type: "preresolved", URL: i.url}}, nil
}
