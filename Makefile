docker-clean-dev:
	docker -v && \
	docker stop $(docker ps -aq); \
	docker rm $(docker ps -aq); \
	rm -r -f data; \
	docker system prune -f; \
