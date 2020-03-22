DOCKER_COMPOSE_FILE		:= "docker-compose.yml"
DOCKERFILE_FILE			:= "../Dockerfile"
DOCKERFILEDEV_FILE		:= "../Dockerfile.dev"

# Builds & run a production-ready image
.PHONY: up
up:
	DOCKERFILE=${DOCKERFILE_FILE} docker-compose up -d --build --scale counter=3

# Builds & run a development-ready image
.PHONY: dev
dev:
	DOCKERFILE=${DOCKERFILEDEV_FILE} docker-compose up -d --scale counter=3

# Run tests inside containers
.PHONY: test
test:
	make dev
	docker-compose exec coordinator sh -c "cd src/coordinator && go test ./... -count=1"
	docker-compose exec counter sh -c "cd src/counter && go test ./... -count=1"

# Removes all containers and all volumes
.PHONY: build-rm-containers
build-rm-containers:
	docker-compose -f ${DOCKER_COMPOSE_FILE} down -v
	docker-compose -f ${DOCKER_COMPOSE_FILE} rm -v

# Stops all the things
.PHONY: stop
stop:
	docker-compose -f ${DOCKER_COMPOSE_FILE} stop

# Clean everything
.PHONY: clean
clean: stop build-rm-containers