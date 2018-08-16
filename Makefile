docker-clean-dev:
	docker -v && \
	docker stop $(docker ps -aq); \
	docker rm $(docker ps -aq); \
	rm -r -f data; \
	docker system prune -f; \

docker-setup-prod:
	cp -vf database.prod.yml database.yml
	cp -vf docker-compose.prod.yml docker-compose.yml

docker-setup-dev:
	cp -vf database.dev.yml database.yml &&
	cp -vf docker-compose.dev.yml docker-compose.yml
