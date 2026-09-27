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

package libgen

import "net/url"

// ClassicSearchMirrors contains the "classic" libgen.rs-derived mirrors.
// As of 2026 this whole family is largely unreachable, but they're kept
// here in case they come back online; GetWorkingMirror() will simply
// skip past them in favor of a working "libgen+" mirror below.
//
// GetDetails (used by the `download <hash>`/`link <hash>` commands) only
// understands the classic json.php API, so those commands deliberately
// restrict mirror selection to this list rather than SearchMirrors.
var ClassicSearchMirrors = []url.URL{
	{
		Scheme: "https",
		Host:   "libgen.is",
		Path:   "search.php",
	},
	{
		Scheme: "https",
		Host:   "libgen.rs",
		Path:   "search.php",
	},
	{
		Scheme: "https",
		Host:   "libgen.st",
		Path:   "search.php",
	},
	{
		Scheme: "https",
		Host:   "libgen.gs",
		Path:   "search.php",
	},
	//{
	//	Scheme: "https",
	//	Host:   "libgen.rocks",
	//	Path:   "index.php",
	//},
	{
		Scheme: "http",
		Host:   "gen.lib.rus.ec",
		Path:   "search.php",
	},
	{
		Scheme: "https",
		Host:   "93.174.95.27",
		Path:   "search.php",
	},
}

// LibgenPlusSearchMirrors contains mirrors running the newer "libgen+"
// codebase (originally libgen.li), which uses a different search endpoint
// and result page layout than the classic mirrors above. As of 2026 this
// is the family that's actually still online.
var LibgenPlusSearchMirrors = []url.URL{
	{
		Scheme: "https",
		Host:   "libgen.li",
		Path:   "index.php",
	},
	{
		Scheme: "https",
		Host:   "libgen.gl",
		Path:   "index.php",
	},
	{
		Scheme: "https",
		Host:   "libgen.la",
		Path:   "index.php",
	},
	{
		Scheme: "https",
		Host:   "libgen.bz",
		Path:   "index.php",
	},
	{
		Scheme: "https",
		Host:   "libgen.vg",
		Path:   "index.php",
	},
}

// SearchMirrors contains all known search mirrors, both libgen+ and
// classic, for display purposes (e.g. `status`). Listed libgen+ first
// since that's the family actually online as of 2026; Search()
// dispatches to the right parsing logic based on which mirror was used.
//
// Actual mirror selection should go through GetWorkingSearchMirror(),
// which prefers libgen+ and only falls back to the classic mirrors
// below if every libgen+ mirror is unreachable.
var SearchMirrors = append(append([]url.URL{}, LibgenPlusSearchMirrors...), ClassicSearchMirrors...)

// DownloadMirrors contains all valid and tested mirrors used for
// downloading content from Library Genesis.
var DownloadMirrors = []url.URL{
	{
		Scheme: "https",
		Host:   "library.lol",
		Path:   "main/",
	},
	{
		Scheme: "https",
		Host:   "libgen.pm",
		Path:   "ads",
	},
}

var UploadMirrors = []url.URL{
	{
		Scheme: "https",
		Host:   "library.bz",
		Path:   "/main/upload",
	},
}

var DbdumpsMirrors = []url.URL{
	{
		Scheme: "https",
		Host:   "data.library.bz",
		Path:   "/dbdumps",
	},
}
