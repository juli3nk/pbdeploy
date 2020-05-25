REPODIR := "/go/src/github.com/juliengk/pbdeploy"

.PHONY: dev
dev:
	docker container run -ti --rm \
		--mount type=bind,src=$$PWD,dst=${REPODIR} \
		--mount type=bind,src=$$HOME/Dev/juliengk/go-git,dst=/go/src/github.com/juliengk/go-git \
		-w ${REPODIR} \
		--name pbdeploy_dev \
		juliengk/dev:go
