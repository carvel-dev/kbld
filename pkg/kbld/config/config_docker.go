package config

type SourceDockerOpts struct {
	Build SourceDockerBuildOpts
}

type SourceDockerBuildOpts struct {
	Target     *string
	Pull       *bool
	NoCache    *bool `json:"noCache"`
	File       *string
	RawOptions *[]string `json:"rawOptions"`
}
