// Copyright © 2020 Ryan Ciehanski <ryan@ciehanski.com>
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
	"log"
	"net/url"
	"os"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/ciehanski/libgen-cli/libgen"
)

var linkCmd = &cobra.Command{
	Use:     "link",
	Short:   "Retrieves and displays the direct download link for a specific resource.",
	Long:    `Retrieves and displays the direct download link for a specific resource.`,
	Example: "libgen link 2F2DBA2A621B693BB95601C16ED680F8",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 1 {
			if err := cmd.Help(); err != nil {
				fmt.Printf("error displaying CLI help: %v\n", err)
			}
			os.Exit(1)
		}
		// Ensure provided entry is valid MD5 hash
		re := regexp.MustCompile(libgen.SearchMD5)
		if !re.MatchString(args[0]) {
			fmt.Printf("Please provide a valid MD5 hash\n")
			os.Exit(1)
		}

		// Get flags
		useIpfs, err := cmd.Flags().GetBool("ipfs-mirrors")
		if err != nil {
			fmt.Printf("error getting ipfs-mirrors flag: %v\n", err)
		}

		fmt.Printf("++ Retrieving download link for: %s\n", args[0])

		// GetDetails only understands the classic json.php API, so
		// hash-based lookups are restricted to classic mirrors.
		searchMirror, err := libgen.GetWorkingMirror(libgen.ClassicSearchMirrors)
		if err != nil {
			fmt.Printf("error finding a working mirror: %v\n", err)
			os.Exit(1)
		}
		bookDetails, err := libgen.GetDetails(&libgen.GetDetailsOptions{
			Hashes:       args,
			SearchMirror: searchMirror,
			Print:        false,
		})
		if err != nil {
			// If error, try a different mirror before exiting.
			var secondaryMirror url.URL
			for attempt := 0; attempt < 5; attempt++ {
				secondaryMirror, err = libgen.GetWorkingMirror(libgen.ClassicSearchMirrors)
				if err != nil {
					log.Fatalf("error finding a working mirror: %v", err)
				}
				if secondaryMirror != searchMirror {
					break
				}
			}
			bookDetails, err = libgen.GetDetails(&libgen.GetDetailsOptions{
				Hashes:       args,
				SearchMirror: secondaryMirror,
				Print:        false,
			})
			if err != nil {
				log.Fatalf("error retrieving results from LibGen API: %v", err)
			}
		}
		book := bookDetails[0]

		if err := libgen.GetDownloadURL(book, useIpfs); err != nil {
			fmt.Printf("error getting download URL: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("%v\n", book.DownloadURL)
	},
}

func init() {
	linkCmd.Flags().BoolP("ipfs-mirrors", "i", false, "enforces libgen-cli to download "+
		"results via IPFS mirrors instead of HTTP(S) mirrors.")
}
