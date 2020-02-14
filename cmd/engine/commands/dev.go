package commands

import (
	"fmt"
	"net/http"

	"github.com/battlesnakeio/engine/cmd/engine/commands/server"
	"github.com/battlesnakeio/engine/version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:     "dev",
	Short:   "dev runs the server, and a dev http server for starting games",
	Version: version.Version,
	Run: func(c *cobra.Command, args []string) {
		go server.RootCmd.Run(c, args)
		go func() {
			mux := http.NewServeMux()
			fs := http.FileServer(http.Dir("board"))
			mux.Handle("/", fs)
			log.Info("board available at http://localhost:3009/")
			if err := http.ListenAndServe(":3009", mux); err != nil {
				fmt.Println("Error while trying to serve board: ", err)
			}
		}()
		mux := http.NewServeMux()
		fs := http.FileServer(http.Dir("public"))
		mux.Handle("/", fs)
		log.Info("dev form available at http://localhost:3010/")
		if err := http.ListenAndServe(":3010", mux); err != nil {
			fmt.Println("Error while trying to serve game form: ", err)
		}
	},
}
