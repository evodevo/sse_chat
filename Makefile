compose=docker-compose -f docker-compose.yml

export compose

.PHONY: run
run: ## runs chat server
	$(compose) up