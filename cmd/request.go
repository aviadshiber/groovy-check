package cmd

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

var allowWrite bool

var requestCmd = &cobra.Command{
	Use:   "request <method> <path>",
	Short: "Raw authenticated Jenkins API call (escape hatch)",
	Long: `Sends an authenticated request to <jenkins_url><path>. GET/HEAD run by
default; anything else requires --allow-write since it can trigger a real
build or config change on a shared Jenkins instance.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		method := strings.ToUpper(args[0])
		path := args[1]

		if method != http.MethodGet && method != http.MethodHead && !allowWrite {
			return fmt.Errorf("refusing %s without --allow-write (this hits a shared Jenkins instance)", method)
		}

		jc := loadJenkinsConfig()
		if err := jc.requireToken(); err != nil {
			return err
		}

		req, err := http.NewRequest(method, strings.TrimRight(jc.URL, "/")+path, nil)
		if err != nil {
			return err
		}
		req.SetBasicAuth(jc.User, jc.Token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("request to %s: %w", jc.URL, err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", string(body))
		if resp.StatusCode >= 400 {
			return fmt.Errorf("HTTP %s", resp.Status)
		}
		return nil
	},
}

func init() {
	requestCmd.Flags().BoolVar(&allowWrite, "allow-write", false, "allow non-GET/HEAD methods")
	rootCmd.AddCommand(requestCmd)
}
