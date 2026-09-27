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

package libgen

import (
	"errors"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// IsLibgenPlusMirror reports whether host belongs to the "libgen+"
// (libgen.li-derived) mirror family, which needs different search/download
// parsing than the classic libgen.rs-derived mirrors.
func IsLibgenPlusMirror(host string) bool {
	for _, m := range LibgenPlusSearchMirrors {
		if strings.EqualFold(m.Hostname(), host) {
			return true
		}
	}
	return false
}

// searchLibgenPlus queries a libgen+ mirror's index.php search endpoint and
// parses Books directly out of the results table. Unlike the classic
// mirrors, there's no separate per-book JSON detail call: title, author,
// publisher, year, language, pages, filesize, extension and md5 are all
// present in the search results page itself.
func searchLibgenPlus(options *SearchOptions) ([]*Book, error) {
	var res int
	switch {
	case options.Results <= 25:
		res = 25
	case options.Results <= 50:
		res = 50
	default:
		res = 100
	}

	q := url.Values{}
	q.Set("req", options.Query)
	q.Set("res", strconv.Itoa(res))

	switch options.SortBy {
	case "id":
		q.Set("order", "f_id")
	case "title":
		q.Set("order", "title")
	case "author":
		q.Set("order", "author")
	case "pub":
		q.Set("order", "publisher")
	case "ext":
		q.Set("order", "extension")
	case "year":
		q.Set("order", "year")
	case "size":
		q.Set("order", "filesize")
	case "lang":
		q.Set("order", "language")
	}
	if options.SortBy != "" {
		if options.SortASC {
			q.Set("ordermode", "asc")
		} else {
			q.Set("ordermode", "desc")
		}
	}

	searchURL := options.SearchMirror
	searchURL.RawQuery = q.Encode()

	b, err := getBody(searchURL.String())
	if err != nil {
		return nil, err
	}

	return parseLibgenPlusResults(string(b), options)
}

// parseLibgenPlusResults extracts Books from a libgen+ search results page.
func parseLibgenPlusResults(response string, options *SearchOptions) ([]*Book, error) {
	tableMatch := regexp.MustCompile(libgenPlusTableReg).FindStringSubmatch(response)
	if len(tableMatch) < 2 {
		return nil, nil
	}

	rows := regexp.MustCompile(libgenPlusRowReg).FindAllStringSubmatch(tableMatch[1], -1)
	titleReg := regexp.MustCompile(libgenPlusTitleReg)
	fileIDReg := regexp.MustCompile(libgenPlusFileIDReg)
	md5Reg := regexp.MustCompile(libgenPlusMd5Reg)
	tagReg := regexp.MustCompile(htmlTagReg)

	var books []*Book
	for _, row := range rows {
		if len(row) < 2 {
			continue
		}
		// Each <td> is a sibling within the row (no nested <td>s), so a
		// plain split gives us the columns in table order: title, author,
		// publisher, year, language, pages, size, extension, mirrors.
		cols := strings.Split(row[1], "<td>")
		if len(cols) < 10 {
			// Header row or otherwise malformed - skip it.
			continue
		}

		md5Match := md5Reg.FindStringSubmatch(cols[9])
		if len(md5Match) < 2 {
			continue
		}
		titleMatch := titleReg.FindStringSubmatch(cols[1])
		if len(titleMatch) < 2 {
			continue
		}

		book := &Book{
			Md5:          md5Match[1],
			Title:        cleanCell(tagReg, titleMatch[1]),
			Author:       cleanCell(tagReg, cols[2]),
			Publisher:    cleanCell(tagReg, cols[3]),
			Year:         cleanCell(tagReg, cols[4]),
			Language:     cleanCell(tagReg, cols[5]),
			Pages:        cleanCell(tagReg, cols[6]),
			Filesize:     cleanCell(tagReg, cols[7]),
			Extension:    cleanCell(tagReg, cols[8]),
			SourceMirror: options.SearchMirror.Hostname(),
		}
		if idMatch := fileIDReg.FindStringSubmatch(cols[1]); len(idMatch) >= 2 {
			book.ID = idMatch[1]
		}
		book.PageURL = fmt.Sprintf("https://%s/ads.php?md5=%s", book.SourceMirror, book.Md5)

		// Apply the same filters GetDetails applies for the classic dialect.
		if options.RequireAuthor && book.Author == "" {
			continue
		}
		if len(options.Extension) > 0 {
			validExtension := false
			for _, ext := range options.Extension {
				if ext == book.Extension {
					validExtension = true
				}
			}
			if !validExtension {
				continue
			}
		}
		if options.Year != 0 {
			y, err := strconv.Atoi(book.Year)
			if err != nil || y != options.Year {
				continue
			}
		}
		if options.SortBy == "year" && (book.Year == "" || book.Year == "0") {
			continue
		}
		if options.Publisher != "" && !strings.Contains(strings.ToLower(book.Publisher), strings.ToLower(options.Publisher)) {
			continue
		}
		if options.Language != "" && !strings.EqualFold(book.Language, options.Language) {
			continue
		}

		if options.Print {
			if err := printDetails(book); err != nil {
				return nil, err
			}
		}

		books = append(books, book)
		if len(books) >= options.Results {
			break
		}
	}

	return books, nil
}

// cleanCell strips HTML tags from a table cell and unescapes any HTML
// entities, returning the trimmed plain-text content.
func cleanCell(tagReg *regexp.Regexp, cell string) string {
	return strings.TrimSpace(html.UnescapeString(tagReg.ReplaceAllString(cell, "")))
}

// getLibgenPlusURL retrieves the download link for a Book found via a
// libgen+ mirror. libgen+ mirrors serve downloads directly (via their own
// ads.php -> get.php redirect page) rather than through library.lol or
// libgen.pm.
func getLibgenPlusURL(book *Book) error {
	queryURL := fmt.Sprintf("https://%s/ads.php?md5=%s", book.SourceMirror, book.Md5)
	book.PageURL = queryURL

	b, err := getBody(queryURL)
	if err != nil {
		return err
	}

	downloadPath := findMatch(libgenPMReg, b)
	if downloadPath == nil {
		return errors.New("no valid libgen+ download URL found")
	}
	book.DownloadURL = fmt.Sprintf("https://%s/%s", book.SourceMirror, string(downloadPath))

	return nil
}
