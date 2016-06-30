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
	"html": "text/html",
	"js":   "text/javascript",
	"jpg":  "image/jpeg",
	"png":  "image/png",
}
