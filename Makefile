.PHONY: tf-init tf-plan tf-apply tf-destroy

TF_RUN = docker run --rm -it --network host -v "$$(pwd):/workspace" -w /workspace --env-file .env hashicorp/terraform:latest

tf-init:
	$(TF_RUN) init

tf-plan:
	$(TF_RUN) plan

tf-apply:
	$(TF_RUN) apply

tf-destroy:
	$(TF_RUN) destroy