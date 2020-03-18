DOCKER_COMPOSE_FILE		:= "docker-compose.yml"
DOCKERFILE_FILE			:= "Dockerfile"
DOCKERFILEDEV_FILE		:= "Dockerfile.dev"

# Builds & run a production-ready image
.PHONY: prod
prod:
	DOCKERFILE=${DOCKERFILE_FILE} docker-compose up -d --build

# Builds & run a development-ready image
.PHONY: dev
dev:
	DOCKERFILE=${DOCKERFILEDEV_FILE} docker-compose up -d --build

# Removes all containes and all volumes
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