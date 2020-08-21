# Run natively

app.install:
	docker-compose up -d db
	npm install

app.run:
	npm run watch \
	| go run service/main.go

test.run:
	docker-compose up -d db_test
	sleep 5
	go test -v ./...
	docker-compose stop db_test

# Run via Docker

app.docker.up:
	docker-compose up -d app
	docker-compose logs -f

app.docker.stop:
	docker-compose stop

app.docker.restart:
	docker-compose stop app
	docker-compose up -d app

app.docker.kill:
	docker-compose down

test.docker.up:
	docker-compose up app_test

test.docker.stop:
	docker-compose stop app_test db_test

# Release from local working directory

app.release:
	# requires following vars to be passed to command: BUILD_TAG, RELEASE_TAG, SSH_KEY, SSH_USER, SSH_HOST, DOCKER_PROJECT_DIR
	# this command can be ignored once CI/CD pipelines are configured and working

	# build front end assets
	docker run --rm -v ${PWD}:/app -w /app node:13.10 npm install && npm run prod

	# build and tag image
	docker image build --rm --tag prediction-league-app:${BUILD_TAG} -f ${PWD}/docker/app.Dockerfile .
	docker tag prediction-league-app:${BUILD_TAG} adamjsc/prediction-league-app:${BUILD_TAG}
	docker tag prediction-league-app:${BUILD_TAG} adamjsc/prediction-league-app:latest

	# push to docker repo
	cat ${PWD}/release/docker-access-token | docker login --username adamjsc --password-stdin
	docker push adamjsc/prediction-league-app:${BUILD_TAG}
	docker push adamjsc/prediction-league-app:latest

	# release
	ssh -i ${SSH_KEY} ${SSH_USER}@${SSH_HOST} '/bin/sh -s' < release.sh ${DOCKER_PROJECT_DIR} ${RELEASE_TAG}

	# clean up locally
	docker image rm prediction-league-app:${BUILD_TAG}
	docker image rm adamjsc/prediction-league-app:${BUILD_TAG}
	docker image rm adamjsc/prediction-league-app:latest
	rm -rf vendor
