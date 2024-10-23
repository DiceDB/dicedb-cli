release:
	git tag 0.0.1
	git push origin 0.0.1
	goreleaser release --clean
