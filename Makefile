.DEFAULT_GOAL := list

BASE_URL ?= http://127.0.0.1:8080
CURL := curl -w '\nstatus: %{http_code}\n'
ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))

.PHONY: list create update delete transfer

list:
	$(CURL) $(BASE_URL)/users

create:
	$(CURL) -X POST $(BASE_URL)/users \
		-H 'Content-Type: application/json' \
		-d '{"name":"$(word 1,$(ARGS))","email":"$(word 2,$(ARGS))"}'

update:
	$(CURL) -X PUT $(BASE_URL)/users/$(word 1,$(ARGS)) \
		-H 'Content-Type: application/json' \
		-d '{"name":"$(word 2,$(ARGS))","email":"$(word 3,$(ARGS))"}'

delete:
	$(CURL) -X DELETE $(BASE_URL)/users/$(word 1,$(ARGS))

transfer:
	$(CURL) -X POST $(BASE_URL)/transfer \
		-H 'Content-Type: application/json' \
		-d '{"from_id":$(word 1,$(ARGS)),"to_id":$(word 2,$(ARGS)),"cents":$(word 3,$(ARGS))}'

# 创建、更新、删除、转账时，后续目标作为参数；其他情况直接按 ID 查询。
ifneq ($(filter create update delete transfer,$(firstword $(MAKECMDGOALS))),)
%:
	@:
else
%:
	$(CURL) $(BASE_URL)/users/$@
endif
