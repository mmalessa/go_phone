include .env

APP_NAME = go_phone

ARM_SSH = ssh $(ARM_USER)@$(ARM_IP)
.DEFAULT_GOAL = go-build

#-----------------------------------------------------------------------------------------------------------------------
ARG := $(word 2, $(MAKECMDGOALS))
%:
	@:
#-----------------------------------------------------------------------------------------------------------------------
#-----------------------------------------------------------------------------------------------------------------------

help: ## Outputs this help screen
	@grep -hE '(^[a-zA-Z0-9_-]+:.*?##.*$$)|(^##)' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}{printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}' | sed -e 's/\[32m##/[33m/'

check-env:
	@if ! [ -f .env ]; then echo ".env - file not found" && return 1; fi

init64: ## Init environment (arm64)
	# https://askubuntu.com/questions/1339558/cant-build-dockerfile-for-arm64-due-to-libc-bin-segmentation-fault
	# @docker run --rm --privileged docker/binfmt:a7996909642ee92942dcd6cff44b9b95f08dad64
	# WORKAROUND
	@docker pull tonistiigi/binfmt:latest
	@docker run --privileged --rm tonistiigi/binfmt --uninstall qemu-*
	@docker run --privileged --rm tonistiigi/binfmt --install arm64
	@docker buildx build --platform $(ARM_PLATFORM) --tag $(APP_NAME)-builder --file .docker/images/app-builder/Dockerfile .

init7: ## Init environment (arm/v7)
	@docker run --rm --privileged docker/binfmt:a7996909642ee92942dcd6cff44b9b95f08dad64
	@docker buildx build --platform $(ARM_PLATFORM) --tag $(APP_NAME)-builder --file .docker/images/app-builder/Dockerfile .

init: init7 ## Init environment (alias)


go-mod-init: # Run go mod init
	@echo 'go mod init...'
	@docker run --rm \
		-v "$(PWD)":/usr/src/$(APP_NAME) \
		-v "$(PWD)/var/go:/go" \
		-v "$(PWD)/var/cache:/root/.cache" \
		--name $(APP_NAME)-builder \
		--platform $(ARM_PLATFORM) \
  		-w /usr/src/$(APP_NAME) \
		$(APP_NAME)-builder:latest \
		rm var -rf && rm go.mod -f && rm go.sum -f && go mod init github.com/mmalessa/$(APP_NAME)

go-mod-tidy: # Run go mod tidy
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
arm-uptime: check-env ## Get uptime from RPI
	@echo "ARM $(ARM_IP) uptime..."
	@ssh $(ARM_USER)@$(ARM_IP) 'uptime'

arm-remove-from-known-hosts: check-env
	@echo "ARM $(ARM_IP) remove ssh-key from known_hosts"
	@ssh-keygen -f ~/.ssh/known_hosts -R "$(ARM_IP)"

arm-authorize: check-env ## (keygen &&) ssh-copy-id
	@echo "ARM $(ARM_IP) authorize... (ssh-keygen, ssh-copy-id)"
	@if ! [ -f ~/.ssh/id_rsa.pub ]; then echo "ssh-keygen" && ssh-keygen; fi
	@ssh-copy-id -f $(ARM_USER)@$(ARM_IP)

# requirements
# ARM - orange PI zero with ARMBIAN jammy
arm-init: check-env ## Init orangePI
	@echo "ARM $(ARM_IP) init all packages and configs"
	@$(ARM_SSH) 'apt update && apt install -y portaudio19-dev libmpg123-0 libmp3lame0'
	@scp "./linux/armbian/usb/usb-mount.sh" $(ARM_USER)@$(ARM_IP):/usr/bin/
	@scp "./linux/armbian/usb/usb-mount@.service" $(ARM_USER)@$(ARM_IP):/etc/systemd/system/
	@scp "./linux/armbian/usb/99-usb.rules" $(ARM_USER)@$(ARM_IP):/etc/udev/rules.d/
	@$(ARM_SSH) 'chmod +x /usr/bin/usb-mount.sh && udevadm control --reload-rules && systemctl daemon-reload'
	@scp "./linux/armbian/dts/powerinfo.dts" $(ARM_USER)@$(ARM_IP):/root/
	@$(ARM_SSH) 'armbian-add-overlay /root/powerinfo.dts'
	
arm-send-bin: check-env ## Send binary and config to RPI
	@echo "Send binary and config to RPI"
	@ssh $(ARM_USER)@$(ARM_IP) 'if ! [ -d ~/bin ]; then mkdir ~/bin; fi'
	scp ./bin/$(APP_NAME) $(ARM_USER)@$(ARM_IP):/usr/bin/$(APP_NAME)

arm-enable-service: check-env ## Enable christmastree service on RPI
	@echo "Enable $(APP_NAME) service on ARM $(ARM_IP)..."
	@scp ./linux/armbian/$(APP_NAME).service $(ARM_USER)@$(ARM_IP):/etc/systemd/system/
	@$(ARM_SSH) 'sudo systemctl enable $(APP_NAME).service'

arm-disable-service: check-env ## Enable christmastree service on RPI
	@echo "Disable $(APP_NAME) service on ARM $(ARM_IP)..."
	@$(ARM_SSH) 'sudo systemctl disable $(APP_NAME).service'
	
arm-start-service: check-env ## Start christmastree service on RPI
	@echo "Start $(APP_NAME) service on ARM $(ARM_IP)..."
	@$(ARM_SSH) 'sudo systemctl start $(APP_NAME).service'

arm-stop-service: check-env ## Stop christmastree service on RPI
	@echo "Stop $(APP_NAME) service on RPI $(ARM_IP)..."
	@$(ARM_SSH) 'sudo systemctl stop $(APP_NAME).service'
	
arm-restart-service: check-env ## Restart christmastree service on RPI
	@echo "Restart $(APP_NAME) service on RPI $(ARM_IP)..."
	@$(ARM_SSH) 'sudo systemctl restart $(APP_NAME).service'
	
arm-down: check-env ## Poweroff RPI
	@echo "Send 'poweroff' to ARM $(ARM_IP)..."
	@$(ARM_SSH) 'sudo poweroff'

arm-console: check-env ## SSH RPI console
	@$(ARM_SSH)