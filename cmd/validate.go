package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type validateResult struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

var validateCmd = &cobra.Command{
	Use:   "validate <Jenkinsfile>",
	Short: "Validate declarative pipeline syntax against a live Jenkins server",
	Long: `Posts the file to <jenkins_url>/pipeline-model-converter/validate — the
official Jenkins-maintained syntax check (see jenkins.io Pipeline Development docs).
This is authoritative for declarative-pipeline structure but, like lint, does not
detect CPS-transform runtime bugs (see the root command's --help for details).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jc := loadJenkinsConfig()
		if err := jc.requireToken(); err != nil {
			return err
		}

		content, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("reading %s: %w", args[0], err)
		}

		client := &http.Client{}

		crumbReq, err := http.NewRequest("GET", strings.TrimRight(jc.URL, "/")+"/crumbIssuer/api/json", nil)
		if err != nil {
			return err
		}
		crumbReq.SetBasicAuth(jc.User, jc.Token)
		crumbResp, err := client.Do(crumbReq)
		if err != nil {
			return fmt.Errorf("fetching crumb from %s: %w", jc.URL, err)
		}
		defer crumbResp.Body.Close()

		var crumbField, crumbValue string
		if crumbResp.StatusCode == http.StatusOK {
			var crumbJSON struct {
				Crumb             string `json:"crumb"`
				CrumbRequestField string `json:"crumbRequestField"`
			}
			if err := json.NewDecoder(crumbResp.Body).Decode(&crumbJSON); err == nil {
				crumbField = crumbJSON.CrumbRequestField
				crumbValue = crumbJSON.Crumb
			}
		}

		var body bytes.Buffer
		w := multipart.NewWriter(&body)
		if err := w.WriteField("jenkinsfile", string(content)); err != nil {
			return err
		}
		if err := w.Close(); err != nil {
			return err
		}

		req, err := http.NewRequest("POST", strings.TrimRight(jc.URL, "/")+"/pipeline-model-converter/validate", &body)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", w.FormDataContentType())
		req.SetBasicAuth(jc.User, jc.Token)
		if crumbField != "" {
			req.Header.Set(crumbField, crumbValue)
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("posting to %s: %w", jc.URL, err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		result := validateResult{
			OK:      resp.StatusCode == http.StatusOK && strings.Contains(string(respBody), "successfully validated"),
			Message: strings.TrimSpace(string(respBody)),
		}

		if jsonOutput {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(result)
		}

		fmt.Println(result.Message)
		if !result.OK {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
