VERSION ?= latest

.PHONY: build-all push-all deploy-all delete-all logs-all


# Build Docker images
build-todo:
	@docker build ./part-3/todo/ -t sathirak/todo:${VERSION}
	@docker build ./part-3/todo-backend/ -t sathirak/todo-backend:${VERSION}

# Push Docker images
push-todo: build-todo
	@docker push sathirak/todo:${VERSION}
	@docker push sathirak/todo-backend:${VERSION}

todo: build-todo push-todo deploy-todo 
	@echo "Todo app built, pushed and deployed"

deploy-todo:
	kubectl apply -f ./part-3/todo/mainfests/
	kubectl apply -f ./part-3/todo-backend/mainfests/

delete-todo:
	@kubectl delete -f ./part-3/todo/mainfests/
	@kubectl delete -f ./part-3/todo-backend/mainfests/

# View logs
logs-todo:
	kubectl logs -l app=todo --tail=100 -f

decrypt:
	@export SOPS_AGE_KEY_FILE=$(PWD)/key.txt && \
	sops -d $(file) > secret.yml

encrypt:
	@export SOPS_AGE_KEY_FILE=$(PWD)/key.txt && \
	sops --encrypt \
	--age age1n76yj5wawhxrcu8ck7324u459t5tph2vcj7gymju85nt2xhm3dzqzl3hxf \
	--encrypted-regex '^(data)$$' \
	$(file) > secret.enc.yml