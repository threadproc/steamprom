build:
	docker buildx build --platform linux/amd64 --push -t registry.notk.ai/steamprom:latest -t registry.notk.ai/steamprom:$$(git rev-parse --short HEAD) .