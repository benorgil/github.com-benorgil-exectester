# Force amd64 arch to prevent m1 apple machines from building for arm
export DOCKER_DEFAULT_PLATFORM:= linux/amd64

.PHONY: all
all: build

.PHONY: build
build:
	go run build_pipeline/*.go



############## For testing execution manually ##############
.PHONY: testrun
testrun:
	go build -buildvcs=false -o build/et
	build/et -o ooooooo -e eeeeeeee -r 3

.PHONY: testrun-long
testrun-long:
	go build -buildvcs=false -o build/et
	build/et -o ooooooo -e eeeeeeee -r 30 -x 3