// Copyright © 2026 Ryan Ciehanski <ryan@ciehanski.com>
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
	"os"

	"github.com/spf13/cobra"

	"github.com/ciehanski/libgen-cli/cmd/libgen-cli/tui"
)

var tuiCmd = &cobra.Command{
	Use:     "tui",
	Short:   "Launch the interactive full-screen libgen-cli.",
	Long:    `Launches a persistent, full-screen terminal UI for searching and downloading resources from Library Genesis.`,
	Example: "libgen tui",
	Run: func(cmd *cobra.Command, args []string) {
		output, err := cmd.Flags().GetString("output")
		if err != nil {
			fmt.Printf("error getting output flag: %v\n", err)
		}
		if err := tui.Run(output); err != nil {
			fmt.Printf("error running tui: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	tuiCmd.Flags().StringP("output", "o", "", "where you want "+
		"libgen-cli to save your downloads.")
}
