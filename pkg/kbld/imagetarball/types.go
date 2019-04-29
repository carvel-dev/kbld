package tarball

type ImageOrImageIndexTarDescriptor struct {
	ImageIndex *ImageIndexTarDescriptor
	Image      *ImageTarDescriptor
}

type ImageIndexTarDescriptor struct {
	Refs    []string
	Images  []ImageTarDescriptor
	Indexes []ImageIndexTarDescriptor

	MediaType string
	Digest    string
	Raw       string
}

type ImageTarDescriptor struct {
	Refs   []string
	Layers []ImageLayerTarDescriptor

	Config   ConfigTarDescriptor
	Manifest ManifestTarDescriptor
}

type ImageLayerTarDescriptor struct {
	MediaType string
	Digest    string
	DiffID    string
	Size      int64
}

type ConfigTarDescriptor struct {
	Digest string
	Raw    string
}

type ManifestTarDescriptor struct {
	MediaType string
	Digest    string
	Raw       string
}
