app.run:
	docker-compose up -d db \
	&& npm install \
	&& npm run watch \
	| go run service/main.go

app.up:
	docker-compose up -d app && docker-compose logs -f

app.stop:
	docker-compose stop

app.restart:
	docker-compose stop app && docker-compose up -d app

app.kill:
	docker-compose down

app.release:
	# requires following vars to be passed to command: BUILD_TAG, RELEASE_TAG, SSH_KEY, SSH_USER, SSH_HOST, HOST_DIR

	# build and tag
	docker image build --rm --tag prediction-league-app:${BUILD_TAG} -f ${PWD}/docker/app.Dockerfile .
	docker tag prediction-league-app:${BUILD_TAG} adamjsc/prediction-league-app:${BUILD_TAG}
	docker tag prediction-league-app:${BUILD_TAG} adamjsc/prediction-league-app:latest

	# push to docker repo
	cat ${PWD}/release/docker-access-token | docker login --username adamjsc --password-stdin
	docker push adamjsc/prediction-league-app:${BUILD_TAG}
	docker push adamjsc/prediction-league-app:latest

	# release
	ssh -i ${SSH_KEY} ${SSH_USER}@${SSH_HOST} '/bin/sh -s' < release.sh ${HOST_DIR} ${RELEASE_TAG}

	# clean up locally
	docker image rm prediction-league-app:${BUILD_TAG}
	docker image rm adamjsc/prediction-league-app:${BUILD_TAG}
	docker image rm adamjsc/prediction-league-app:latest
	rm -rf vendor

test.run:
	docker-compose up -d db_test && go test -v ./...

test.up:
	docker-compose up app_test

test.stop:
	docker-compose stop app_test db_test
