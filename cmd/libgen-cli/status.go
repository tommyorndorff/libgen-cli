// Copyright © 2019 Ryan Ciehanski <ryan@ciehanski.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package libgen_cli

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/ciehanski/libgen-cli/libgen"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Checks the status of Library Genesis' mirrors.",
	Long:    `Checks the status of all Library Genesis search mirrors as well as all download mirrors.`,
	Example: `libgen status`,
	Run: func(cmd *cobra.Command, args []string) {

		// Don't allow args
		if len(args) != 0 {
			if err := cmd.Help(); err != nil {
				fmt.Printf("error displaying CLI help: %v\n", err)
			}
			os.Exit(0)
		}

		// Get flags
		mirror, err := cmd.Flags().GetString("mirror")
		if err != nil {
			fmt.Printf("error getting mirror flag: %v\n", err)
		}

		switch mirror {
		case "download":
			printMirrorStatuses(libgen.DownloadMirrors)
		case "search":
			printMirrorStatuses(libgen.SearchMirrors)
		default:
			printMirrorStatuses(libgen.SearchMirrors)
			printMirrorStatuses(libgen.DownloadMirrors)
		}
	},
}

// printMirrorStatuses checks and prints the [OK]/[FAIL] status of each
// mirror in urls.
func printMirrorStatuses(urls []url.URL) {
	for _, u := range urls {
		status := libgen.CheckMirror(u)
		label := color.RedString("[FAIL]")
		if status == http.StatusOK {
			label = color.GreenString("[OK]")
		}
		if runtime.GOOS == "windows" {
			if _, err := fmt.Fprintf(color.Output, "%s %s\n", label, u.Host); err != nil {
				fmt.Printf("error writing to Windows os.Stdout: %v\n", err)
			}
		} else {
			fmt.Printf("%s %s\n", label, u.Host)
		}
	}
}

func init() {
	statusCmd.Flags().StringP("mirror", "m", "", "Choose a specific "+
		"collection of mirrors to check status.")
}
