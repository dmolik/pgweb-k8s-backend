package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dmolik/pgweb-k8s-backend/api"
	"github.com/dmolik/pgweb-k8s-backend/mgr"

	badger "github.com/dgraph-io/badger/v4"
)

func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Run: func(cmd *cobra.Command, args []string) {
		opts := badger.DefaultOptions("").WithInMemory(true)
		opts  = opts.WithIndexCacheSize(100 << 19) // 50MB
		badgerDB, err := badger.Open(opts)
		defer badgerDB.Close()
		if err != nil {
			panic(err)
		}
		go mgr.Start(badgerDB)
		api.Serve(badgerDB)
	},
}
