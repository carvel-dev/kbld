package maven

import (
	ctlbdk "carvel.dev/kbld/pkg/kbld/builder/docker"
	"carvel.dev/kbld/pkg/kbld/config"
	ctllog "carvel.dev/kbld/pkg/kbld/logger"
)

type Jib struct {
	docker ctlbdk.Docker
	logger ctllog.Logger
}

func NewMavenJib(docker ctlbdk.Docker, logger ctllog.Logger) Jib {
	return Jib{docker: docker, logger: logger}
}

func (b *Jib) Run(image, directory string, opts config.SourceJibRunOpts) (ctlbdk.TmpRef, error) {

	// TODO add logic to run maven jib build command

	return ctlbdk.TmpRef{}, nil
}
