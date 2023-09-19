default: help

##@
cluster-create:
	sh ./tilt/scripts/cluster-create.sh
.PHONY: cluster-create

cluster-destroy:
	sh ./tilt/scripts/assert-context.sh
	sh ./tilt/scripts/cluster-destroy.sh
.PHONY: cluster-destroy
