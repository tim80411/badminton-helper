# 變數定義
IMAGE_NAME := badminton-helper
IMAGE_TAG := latest
DOCKER_REGISTRY ?=
FULL_IMAGE_NAME := $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/$(IMAGE_NAME),$(IMAGE_NAME))

.PHONY: help build-image run-image push-image clean-image

# 預設目標：顯示幫助訊息
help:
	@echo "可用的指令："
	@echo "  make build-image    - 建置 Docker image"
	@echo "  make run-image      - 執行 Docker container"
	@echo "  make push-image     - 推送 image 到 registry"
	@echo "  make clean-image    - 清除本地 Docker image"

# 建置 Docker image
build-image:
	@echo "正在建置 Docker image..."
	docker build -t $(FULL_IMAGE_NAME):$(IMAGE_TAG) .
	@echo "建置完成: $(FULL_IMAGE_NAME):$(IMAGE_TAG)"

# 執行 Docker container
run-image:
	@echo "正在啟動 Docker container..."
	docker run -d \
		--name $(IMAGE_NAME) \
		-p 8080:8080 \
		-v $(PWD)/credentials:/home/appuser/credentials \
		$(FULL_IMAGE_NAME):$(IMAGE_TAG)
	@echo "Container 已啟動，可透過 http://localhost:8080 存取"

# 推送 image 到 registry
push-image:
	@if [ -z "$(DOCKER_REGISTRY)" ]; then \
		echo "錯誤: 請設定 DOCKER_REGISTRY 變數"; \
		echo "範例: make push-image DOCKER_REGISTRY=your-registry.com"; \
		exit 1; \
	fi
	@echo "正在推送 image 到 registry..."
	docker push $(FULL_IMAGE_NAME):$(IMAGE_TAG)
	@echo "推送完成"

# 清除本地 Docker image
clean-image:
	@echo "正在清除 Docker image..."
	docker rmi $(FULL_IMAGE_NAME):$(IMAGE_TAG) || true
	@echo "清除完成"
