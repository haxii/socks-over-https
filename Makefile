VERSION=$(shell git describe --abbrev=0 --tags)

.PHONY: release
release:
	gox -osarch="linux/amd64 linux/386 darwin/amd64 windows/amd64" -output="./dist/{{.Dir}}-$(VERSION)-{{.OS}}-{{.Arch}}/{{.Dir}}"
	cd dist; ls | while read line; do echo $${line}; cp ../README.md ../config.json $${line}; tar cvf ../$${line}.tar.gz $${line}; done

.PHONY: clean
clean:
	rm -rf dist
