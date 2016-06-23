/*
 *
 *
 * Copyright 2012-2016 Viant.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 *  use this file except in compliance with the License. You may obtain a copy of
 *  the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 *  License for the specific language governing permissions and limitations under
 *  the License.
 *
 */

// Package toolbox - mime types
package toolbox

const (
	//JSONMimeType JSON  mime type constant
	JSONMimeType = "text/json"
	//CSVMimeType csv  mime type constant
	CSVMimeType = "text/csv"
	//TSVMimeType tab separated mime type constant
	TSVMimeType = "text/tsv"
	//TextMimeType mime type constant
	TextMimeType = "text/sql"
)

//FileExtensionMimeType json, csv, tsc, sql mime types.
var FileExtensionMimeType = map[string]string{
	"json": JSONMimeType,
	"csv":  CSVMimeType,
	"tsv":  TSVMimeType,
	"sql":  TextMimeType,
	"html":"text/html",
	"js":"text/javascript",
	"jpg":"image/jpeg",
	"png":"image/png",
}
