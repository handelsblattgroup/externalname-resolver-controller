package watch

import (
	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/cli/watch/options"
	"github.com/handelsblattgroup/externalname-resolver-controller/pkg/kube/watcher"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var rootCmd = &cobra.Command{
	Use:   "watch",
	Short: "start the controller",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		if options.Current.Debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}

		watcher, err := watcher.New(options.Current)
		if err != nil {
			return errors.Wrapf(err, "could not initialise watcher")
		}

		log.Debug().Msg("This message appears only when log level set to Debug")
		log.Info().Msg("starting to watch")
		watcher.Watch()

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&options.Current.Kubeconfig, "kubeconfig", options.Current.Kubeconfig, "path of kubeconfig to use. the environment variable KUBECONFIG has higher priority")
	rootCmd.PersistentFlags().DurationVar(&options.Current.ResyncInterval, "resync-interval", options.Current.ResyncInterval, "time interval between full resyncs")
	rootCmd.PersistentFlags().BoolVar(&options.Current.Debug, "debug", options.Current.Debug, "output debugging logs")
	rootCmd.PersistentFlags().StringSliceVarP(&options.Current.IgnoreLabels, "ignor-label", "i", options.Current.IgnoreLabels, "a label that will not be copied to the generated Endpoints")
	rootCmd.PersistentFlags().StringVar(&options.Current.ClusterDnsIP, "cluster-dns", options.Current.ClusterDnsIP, "ip of the cluster DNS service")
}

func Command() *cobra.Command {
	pflag.CommandLine.AddFlagSet(rootCmd.Flags())
	return rootCmd
}
