package cmd

import (
	"log"

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
		// setup badgerDB
		opts := badger.DefaultOptions("").WithInMemory(true)
		opts  = opts.WithIndexCacheSize(100 << 19) // 50MB
		badgerDB, err := badger.Open(opts)
		defer func(){
			if err := badgerDB.Close(); err != nil {
				log.Fatal(err)
			}
		}()
		if err != nil {
			log.Fatal(err)
		}

		// start the controllers
		go mgr.Start(badgerDB)
		// start the API server
		api.Serve(badgerDB)
	},
}
