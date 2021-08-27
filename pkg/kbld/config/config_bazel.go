package config

type SourceBazelOpts struct {
	Build SourceBazelBuildOpts
}

type SourceBazelBuildOpts struct {
	Label      *string   `json:"label"`
	RawOptions *[]string `json:"rawOptions"`
}
