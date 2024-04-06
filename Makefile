BINARY_NAME=codeduel-lobby.exe
DOCKERHUB_USERNAME=xedom
DOCKER_IMAGE_NAME=codeduel-lobby
DOCKER_CONTAINER_NAME=codeduel-lobby
PORT=5010
ENV_FILE=.env.docker

build:
	go build -o ./bin/$(BINARY_NAME) -v

run: build
	./bin/$(BINARY_NAME)

dev:
	go run .

test:
	go test -v ./...

docker-build:
	docker build -t $(DOCKERHUB_USERNAME)/$(DOCKER_IMAGE_NAME) .

docker-push:
	docker push $(DOCKERHUB_USERNAME)/$(DOCKER_IMAGE_NAME)

# docker run -d -p $(PORT):$(PORT) -v $(PWD)\.env.docker:/.env --name $(DOCKER_CONTAINER_NAME) $(DOCKERHUB_USERNAME)/$(DOCKER_IMAGE_NAME)
docker-up:
	docker run -d -p $(PORT):$(PORT) --name $(DOCKER_CONTAINER_NAME) --env-file $(ENV_FILE) $(DOCKERHUB_USERNAME)/$(DOCKER_IMAGE_NAME)

docker-down:
	docker stop $(DOCKER_CONTAINER_NAME)
	docker rm $(DOCKER_CONTAINER_NAME)

docker-restart: docker-down docker-up

release:
	git checkout release
	git merge main
	git push origin release
	git checkout main

clean:
	go clean
	rm -f bin/$(BINARY_NAME)
