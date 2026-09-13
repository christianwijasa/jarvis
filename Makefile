.PHONY: help setup setup-status

help:
	@echo "  make setup         clone missing + pull existing manifest repos"
	@echo "  make setup-status  show clone state without changing anything"

setup:
	./scripts/setup-workstation.sh sync

setup-status:
	./scripts/setup-workstation.sh status
