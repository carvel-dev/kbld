package config

type SourcePackOpts struct {
	Build SourcePackBuildOpts
}

type SourcePackBuildOpts struct {
	Builder    *string
	Buildpacks *[]string
	ClearCache *bool     `json:"clearCache"`
	RawOptions *[]string `json:"rawOptions"`
}
