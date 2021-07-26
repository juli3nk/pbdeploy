REPODIR := "/go/src/github.com/juli3nk/pbdeploy"
IMAGE_NAME := juli3nk/pbdeploy

.PHONY: dev
dev:
	docker container run -ti \
		--rm \
		--mount type=bind,src=$$PWD,dst=${REPODIR} \
		--mount type=bind,src=$$HOME/Dev/juli3nk/go-git,dst=/go/src/github.com/juli3nk/go-git \
		-w ${REPODIR} \
		--name pbdeploy_dev \
		juli3nk/dev:go

.PHONY: build
build:
	docker image build -t $(IMAGE_NAME) .

.PHONY: push
push:
	docker image push $(IMAGE_NAME)
