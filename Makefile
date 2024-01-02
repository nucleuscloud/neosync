default: help

##@
cluster-create:
	sh ./tilt/scripts/cluster-create.sh
.PHONY: cluster-create

cluster-destroy:
	bash ./tilt/scripts/assert-context.sh
	sh ./tilt/scripts/cluster-destroy.sh
.PHONY: cluster-destroy

all: 
	sh -c "cd ./worker && make build"
	sh -c "cd ./backend && make all"
	sh -c "cd ./cli && make build"
