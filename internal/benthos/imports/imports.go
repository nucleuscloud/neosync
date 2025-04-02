package benthos_imports

import (
	_ "github.com/redpanda-data/benthos/v4/public/components/io"
	_ "github.com/redpanda-data/benthos/v4/public/components/pure"
	_ "github.com/redpanda-data/benthos/v4/public/components/pure/extended"

	_ "github.com/redpanda-data/connect/v4/public/components/aws"
	_ "github.com/redpanda-data/connect/v4/public/components/gcp"

	_ "github.com/nucleuscloud/neosync/worker/pkg/benthos/transformers"
)
