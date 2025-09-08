.PHONY: test test-coverage lint

test:
	go test -v ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	COLOR="red"; \
	if [ "$$(echo "$$COVERAGE >= 80" | bc -l 2>/dev/null || echo "0")" = "1" ]; then COLOR="brightgreen"; \
	elif [ "$$(echo "$$COVERAGE >= 60" | bc -l 2>/dev/null || echo "0")" = "1" ]; then COLOR="yellow"; \
	elif [ "$$(echo "$$COVERAGE >= 40" | bc -l 2>/dev/null || echo "0")" = "1" ]; then COLOR="orange"; fi; \
	sed -i.bak "s/Coverage-[0-9]*%25-[a-z]*/Coverage-$${COVERAGE}%25-$$COLOR/" README.md && rm README.md.bak; \
	echo "Coverage report generated: coverage.html ($$COVERAGE%)"

lint:
	golangci-lint run