package imports

import (
	"github.com/phantom-d/go-daemons/imports"
)

func init() {
	imports.Factory.Register("ProductPrices", func() imports.WorkerInterface { return &ProductPrices{} })
}
