docker-clean-dev:
	docker -v && \
	docker stop $(docker ps -aq); \
	docker rm $(docker ps -aq); \
	rm -r -f data; \
	docker system prune -f; \

docker-clean-compose:
	docker-compose rm --all; \
	docker-compose pull; \
	docker-compose build --no-cache; \
	docker-compose up -d --force-recreate;

docker-setup-prod:
	cp -vf database.prod.yml database.yml
	cp -vf docker-compose.prod.yml docker-compose.yml

docker-setup-dev:
	cp -vf database.dev.yml database.yml
	cp -vf docker-compose.dev.yml docker-compose.yml
