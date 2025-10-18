.PHONY: migrate import-fixtures dev
migrate:
	@mkdir -p data
	@echo "Apply schema.sql (stub) — real migrations will be wired in Plan 1"

import-fixtures:
	@echo "Importer stub — will load fixtures after we add cmd/import"

dev: migrate import-fixtures
