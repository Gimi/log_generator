IMAGE_NAME?=genlog
IMAGE_VERSION?=0.0.1
HUB_REPO?=gimil

genlog:
	GOOS=linux GOARCH=amd64 go build -o genlog ./main.go


docker: Dockerfile genlog
	docker build -t ${IMAGE_NAME}:${IMAGE_VERSION} .


tag: docker
	docker tag ${IMAGE_NAME}:${IMAGE_VERSION} ${HUB_REPO}/${IMAGE_NAME}:${IMAGE_VERSION}


push: tag
	docker push ${HUB_REPO}/${IMAGE_NAME}:${IMAGE_VERSION}
