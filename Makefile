default:
	@$(RM) ./game/site; make demo
.PHONY: default

demo:
	@cd game; make site && env HTTP_DEMO=simple-ui ./site
.PHONY: demo

help:
	@echo '--- try: make iGod' # or mango.{app,lib} etc.
.PHONY: help

iGod:
	@make -C gpt
.PHONY: iGod
