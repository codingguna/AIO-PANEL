package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/codingguna/aio-panel/internal/discovery"
	"github.com/codingguna/aio-panel/internal/webserver"
	"github.com/spf13/cobra"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Manage Nginx virtual hosts, reverse proxies, and SSL certificates",
}

var webVHostsCmd = &cobra.Command{
	Use:   "vhosts",
	Short: "List all discovered and AIO-managed virtual hosts",
	RunE: func(cmd *cobra.Command, args []string) error {
		sites, err := discovery.DiscoverNginxSites(context.Background())
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "DOMAIN\tTARGET / ROOT\tSSL\tOWNER\tCONFIG FILE")
		fmt.Fprintln(w, "------\t-------------\t---\t-----\t-----------")

		for _, s := range sites {
			target := s.ProxyPass
			if target == "" {
				target = s.DocumentRoot
			}
			sslStr := "❌ No"
			if s.SSL {
				sslStr = "🔒 Yes"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				s.Domain, target, sslStr, s.OwnerType, s.ConfigFile)
		}
		w.Flush()
		fmt.Println("")
		return nil
	},
}

var webSSLCmd = &cobra.Command{
	Use:   "ssl",
	Short: "List TLS/SSL certificates and expiration status",
	RunE: func(cmd *cobra.Command, args []string) error {
		certs, err := webserver.ListCertificates(context.Background())
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "DOMAIN\tISSUER\tDAYS REMAINING\tAUTO RENEW\tEXPIRATION DATE")
		fmt.Fprintln(w, "------\t------\t--------------\t----------\t---------------")

		for _, c := range certs {
			fmt.Fprintf(w, "%s\t%s\t%d days\t%v\t%s\n",
				c.Domain, c.Issuer, c.DaysRemaining, c.AutoRenew, c.ValidTo.Format("2006-01-02"))
		}
		w.Flush()
		fmt.Println("")
		return nil
	},
}

func init() {
	webCmd.AddCommand(webVHostsCmd)
	webCmd.AddCommand(webSSLCmd)
	RootCmd.AddCommand(webCmd)
}
