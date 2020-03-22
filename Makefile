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

.PHONY: log
log:
	docker-compose logs coordinator counter

# Run simulation
.PHONY: simulate
simulate:
	make clean dev
	# 1. counters should be added and items should be empty
	sleep 10s && docker logs coordinator && curl localhost:8080/items/test/count
	# 2. item should be added
	curl localhost:8080/items -d '[{"id":"item1","tenant":"test"}]' && curl localhost:8080/items/test/count
	# 3. first counter should not respond
	docker stop distributed-counter_counter_1 && sleep 10s && docker logs coordinator
	# 4. first counter should be recovered
	docker start distributed-counter_counter_1 && sleep 10s && docker logs coordinator
	# 5. second and third counter should be marked as not query able
	docker stop distributed-counter_counter_2 distributed-counter_counter_3 && sleep 10s && docker logs coordinator
	# 6. more items should be added
	curl localhost:8080/items -d '[{"id":"item2","tenant":"test"}, {"id":"item3","tenant":"test"}]'
	curl localhost:8080/items/test/count
	# 7. new counters should be added and should imediately have valid items
	docker start distributed-counter_counter_2 distributed-counter_counter_3
	docker-compose exec coordinator sh -c "curl distributed-counter_counter_1/items/test/count && curl distributed-counter_counter_2/items/test/count && curl distributed-counter_counter_3/items/test/count"
	make clean
