.DEFAULT_GOAL := list

BASE_URL ?= http://127.0.0.1:8080
ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))

.PHONY: list create update delete

list:
	curl $(BASE_URL)/users

create:
	curl -X POST $(BASE_URL)/users \
		-H 'Content-Type: application/json' \
		-d '{"name":"$(word 1,$(ARGS))","email":"$(word 2,$(ARGS))"}'

update:
	curl -X PUT $(BASE_URL)/users/$(word 1,$(ARGS)) \
		-H 'Content-Type: application/json' \
		-d '{"name":"$(word 2,$(ARGS))","email":"$(word 3,$(ARGS))"}'

delete:
	curl -X DELETE $(BASE_URL)/users/$(word 1,$(ARGS))

# 创建、更新、删除时，后续目标作为参数；其他情况直接按 ID 查询。
ifneq ($(filter create update delete,$(firstword $(MAKECMDGOALS))),)
%:
	@:
else
%:
	curl $(BASE_URL)/users/$@
endif
