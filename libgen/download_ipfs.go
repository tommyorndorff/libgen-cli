// Copyright © 2023 Ryan Ciehanski <ryan@ciehanski.com>
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

package libgen

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"github.com/cheggaaa/pb/v3"
)

// DownloadBookIPFS downloads the book via its IPFS gateway DownloadURL
// (e.g. https://gateway.ipfs.io/ipfs/<cid>?filename=...). This is a plain
// HTTP(S) fetch against the gateway, since libgen's IPFS links already
// resolve through an HTTP gateway rather than requiring a local IPFS node.
func DownloadBookIPFS(book *Book, outputPath string) error {
	filename := getBookFilename(book)

	req, err := http.NewRequest("GET", book.DownloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept-Encoding", "*")
	client := http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}
	r, err := client.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to reach IPFS gateway %v: HTTP %v", req.Host, r.StatusCode)
	}

	bar := pb.Full.Start64(r.ContentLength)

	out, err := makeFile(outputPath, filename)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, bar.NewProxyReader(r.Body)); err != nil {
		return err
	}

	bar.Finish()

	return nil
}
