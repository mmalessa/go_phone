include .env

APP_NAME = go_phone

.DEFAULT_GOAL = go-build

#-----------------------------------------------------------------------------------------------------------------------
ARG := $(word 2, $(MAKECMDGOALS))
%:
	@:
#-----------------------------------------------------------------------------------------------------------------------
#-----------------------------------------------------------------------------------------------------------------------

help: ## Outputs this help screen
	@grep -hE '(^[a-zA-Z0-9_-]+:.*?##.*$$)|(^##)' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}{printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}' | sed -e 's/\[32m##/[33m/'


init: ## Init environment (add arm support && build ws2811-builder)
	# https://askubuntu.com/questions/1339558/cant-build-dockerfile-for-arm64-due-to-libc-bin-segmentation-fault
	# @docker run --rm --privileged docker/binfmt:a7996909642ee92942dcd6cff44b9b95f08dad64
	# WORKAROUND
	@docker pull tonistiigi/binfmt:latest
	@docker run --privileged --rm tonistiigi/binfmt --uninstall qemu-*
	@docker run --privileged --rm tonistiigi/binfmt --install arm64
	@docker buildx build --platform $(ARM_PLATFORM) --tag $(APP_NAME)-builder --file .docker/images/app-builder/Dockerfile .

go-modtidy: # Run go mod tidy
	@echo 'go mod tidy...'
	@docker run --rm \
		-v "$(PWD)":/usr/src/$(APP_NAME) \
		-v "$(PWD)/var/go:/go" \
		-v "$(PWD)/var/cache:/root/.cache" \
		--name $(APP_NAME)-builder \
		--platform $(ARM_PLATFORM) \
  		-w /usr/src/$(APP_NAME) \
		$(APP_NAME)-builder:latest \
		go mod tidy -v

go-build: ## Run go build
	@echo 'go build -o "bin/$(APP_NAME)"...'
	@docker run --rm \
		-v "$(PWD)":/usr/src/$(APP_NAME) \
		-v "$(PWD)/var/go:/go" \
		-v "$(PWD)/var/cache:/root/.cache" \
		--name $(APP_NAME)-builder \
		--platform $(ARM_PLATFORM) \
  		-w /usr/src/$(APP_NAME) \
		$(APP_NAME)-builder:latest \
		env CGO_ENABLED=1 go build -o "bin/$(APP_NAME)" -v

clean: ## Remove binary from bin/ directory
	@rm -rf bin/${APP_NAME}

## ARM commands
arm-uptime: ## Get uptime from RPI
	@echo "ARM $(ARM_IP) uptime..."
	@ssh $(ARM_USER)@$(ARM_IP) 'uptime'

arm-authorize: ## (keygen &&) ssh-copy-id
	@echo "ARM $(ARM_IP) authorize... (ssh-keygen, ssh-copy-id)"
	@if ! [ -f ~/.ssh/id_rsa.pub ]; then echo "ssh-keygen" && ssh-keygen; fi
	@ssh-copy-id -f $(ARM_USER)@$(ARM_IP)

arm-install: ## Send binary and config to RPI
	@echo "Send binary and config to RPI"
	@ssh $(ARM_USER)@$(ARM_IP) 'if ! [ -d ~/bin ]; then mkdir ~/bin; fi'
	scp ./bin/$(APP_NAME) $(ARM_USER)@$(ARM_IP):~/bin/$(APP_NAME)
	## @ssh $(ARM_USER)@$(ARM_IP) 'rm ~/bin/config/0*.yaml -f'
	## scp ./config/* $(ARM_USER)@$(ARM_IP):~/bin/config/


arm-enable-service: ## Enable christmastree service on RPI
	@echo "Enable $(APP_NAME) service on ARM $(ARM_IP)..."
	@echo "TODO"
	
arm-start-service: ## Start christmastree service on RPI
	@echo "Start $(APP_NAME) service on ARM $(ARM_IP)..."
	@echo "TODO"

arm-stop-service: ## Stop christmastree service on RPI
	@echo "Start $(APP_NAME) service on RPI $(ARM_IP)..."
	@echo "FIXME"
	
arm-restart-service: ## Restart christmastree service on RPI
	@echo "Restart $(APP_NAME) service on RPI $(ARM_IP)..."
	@echo "FIXME"
	
arm-down: ## Poweroff RPI
	@echo "Send 'poweroff' to ARM $(ARM_IP)..."
	@ssh $(ARM_USER)@$(ARM_IP) 'sudo poweroff'

arm-console: ## SSH RPI console
	@ssh $(ARM_USER)@$(ARM_IP)