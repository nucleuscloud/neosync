default: help

##@
cluster-create:
	sh ./tilt/scripts/assert-context.sh
	ctlptl apply -f tilt/kind/cluster.yaml
.PHONY: cluster-create

cluster-destroy:
	sh ./tilt/scripts/assert-context.sh
	ctlptl delete -f tilt/kind/cluster.yaml
.PHONY: cluster-destroy
