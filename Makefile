docker-clean-dev:
	docker -v && \
	docker stop $(docker ps -aq); \
	docker rm $(docker ps -aq); \
	rm -r -f data; \
	docker system prune -f; \

docker-setup-prod:
	cp -f database.prod.yml database.yml
	cp -f docker-compose.prod.yml docker-compose.yml

docker-setup-dev:
	cp -f database.dev.yml database.yml
	cp -f docker-compose.dev.yml docker-compose.yml
