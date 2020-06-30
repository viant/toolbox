## June 30 2020 - v.33.0
    - Added sampler
    
## April 20 2020 - v.31.2
    - Added ToLower, ToUpper in data.udf
    - Add AsNewLineDelimitedJSON in data.udf

## March 31 2020 - v.31.0
    - Added bridge.ReadRecordedHttpTripsWithTemplate
    
## Fab 21 2020  - v.30.0
    - Added CopyMap
    - Added OmitEmptyMapWriter

## Feb 7 2020 - v0.29.6
    - Added data.udf.FormatTime with trunc parameter
    - Added cred.Config.Endpoint
    - Added http options to RouteToService
    
## October 9 - v0.29.0
- Added data.udf.Replace

## September 23 - v0.27.1
- Added Body matcher
- Added url.Resource CustomKey

## July 29 2019 - v0.26.2
- Extended cred.Config
- Patched gs.SetProvider

## July 26 2019 - v0.26.0
- Removed Base64 flag from kms
- Added base64 auto detection
- Added credConfig provider for gs
- Added kms aws service 
- Add url.Resource: DownloadBase64


## July 20 2019 - v0.25.2
- Update supported secret sequences 

## July 18 2019 - v0.25.0
- Add Storage.UploadWithMode
- Patched/streamlined SSH/SCP Upload
- Streamline storage.Copy 

## June 10 2019 - v0.24.0

- Added IsStructuredJSON
- Updated IsCompleteJSON to be compatible with json.Valid - this is breaking change if you need check is input is structure json use IsStructuredJSON
    

## June 7 2019 - v0.23.8
- add Split udf (data/udf)

## June 1 2019 - v0.23.7
- patched TimeAt.Next

## May 22 2019 - v0.23.0
- Renamed StorageProvider to Registry

## May 21 2019 - v0.22.0
- Added AtTime schedule type
- Minor patches

## May 13 2019 - v0.21.1
- Patched AsString

## May 7 2019 - v0.21.0
- Merge KMS helper

## May 6 2019 - v0.20.3
- Patched yaml detection

## April 17 2019 - v0.20.2
- Added cred.Config Email field
- Patched zero value access error

## April 17 2019 - v0.20.0
- Added paraphrase ssh key support
- Added AsFunctionParameters
- Added string to []string conversion path


## April 14 2019 - v0.19.4
- Added StructFields sort type

## April 12 2019 - v0.19.3
- updated ReclassifyNotFoundIfMatched

## March 20 2019 - v0.19.1
- Added SplitTextStream util function

## March 20 2019 - v0.19.0
- Added gs storage with JSON credentials

## March 14 2019 - v0.19.0
 - Added storage Tar archive support
 - Added IsDirectory helper method
 - Patched filesetinfo (function parameters names)
 - Added FileInfo to FileInfo to access relevant imports
 
## March 12 2019 - v0.18.4
 - Patch json dependencies
 
## March 8 2019 - v0.18.3
 - Patched storage scp file parsing

## March 6 2019 - v0.18.2
 - Added AsStringMap udf

## March 1 2019 - v0.18.1
 - Added custom converter Registry with RegisterConverter and GetConverter functions
 - Added TimeWindow util 

## Feb 23 2019 - v0.18.0
 - Added Fields and Ranger method on Compacted slice
 - Made compacted slice field public
 - Added Rand udf, patched udf mappings
 - Patched $AsString udf
 - Updated AsSlice to support Iterator
 
## Feb 13 2019 - v0.17.0
 - Added $Select UDF 

## Feb 12 2019 - v0.16.2
 - patched *number nil pointer error
 - patched anonymous time struct conversion

## Feb 10 2019 - v0.16.0
 - Added ConstValueProvider
 - Patched BuildTagMapping with anonymous explicit JSON field
 - Enhanced cast value provided with int32, int64, float32
 - Patched NormalizeKVPairs

 
## Feb 3 2019 - v0.15.1
 - Added Slice Intersect to collections 

## Feb 3 2019 - v0.15.11
 - Added Intersect utility method
 - Added Sum, Count, AsNumber Elapsed data/udf
 - Patched StructHelper
 - Patched Converter
 

## Feb 3 2019 - v0.14.0
  - Added ToCaseFormat text util
  - Updated fileset reader to read interface method
  - Patched struct_helper panic
  - Optimized ReverseSlice
 
## Jan 31 2019 - v0.12.0
  - Added storage/copy:ArchiveWithFilter
  - Added default project detection to google storage
    
## Jan 30 2019 - v0.11.1
  - Patched new line delimited json decoding
  - Patched conversion slice to map error handling
  
## Jan 29 2019 - v0.11.0
  - Remove storage/aws package, use storage/s3 instead - breakable change
  - Added google storage default http clinet (to run within GCE, or with GOOGLE_APPLICATION_CREDENTIALS)  
  - Added google storage customization with GOOGLE_STORAGE_PROJECT env variable 
  - Patched nil pointer check on fileset_info
    
## Jan 24 2019 - v0.10.3
  - Update file set info fix IsPointerComponent in slice component type
  - Added recursive remove on file storage impl 
  

## Jan 23 2019 - v0.10.1
  - Added MaxIdleConnsPerHost http client option
  - Added TrimSpaces data/udf

## Jan 19 2019 - v0.10.0
  - Added IsNumber helper function
  - Enhance Process Struct to handle unexported fields
    * Added UnexportedFieldHandler hadnler
    * Defined SetUnexportedFieldHandler function
    * Defined IgnoreUnexportedFields default handler
  - Enhanced GetStructMeta to handle unexported fields
    * Added StructMetaFilter
    * Defined DefaultStructMetaFilter default
    * Defined SetStructMetaFilter


## Jan 15 2019 - v0.9.0
   - DownloadWithURL(URL string) (io.ReadCloser, error)
   
## Jan 13 2019 - v0.8.1
  - Moved storage/aws to storage/s3
  - Added lazy s3 bucket creation on upload
  - Added data/udf QueryUnescape 


## Jan 13 2019 - v0.7.0
  - Added ScanStructMethods method
  - Added TryDiscoverValueByKind method
  - Patched AsString UDF
  - Added Base64DecodeText udf
  - Added AnyJSONType for generic interface{} types
  - Added AccountID to cred/config
  - Patched []uint data substitution parsing
    
## Jan 8 2019 - v0.6.5
 - Updated struct to map conversion with honoring tag name

## Jan 7 2019 - v0.6.5
 - Patched non writable map data type mutation

## Jan 6 2019 - v0.6.4
 - Added nested array mutation in data.Map
 - Patched url.resource yaml loading with array structure
 - Patched IndexOf
 - Minor patched
 
## Jan 3 2019 - v0.6.3
 - Added NotFound error with helper functions
 - Updated handling not found on upload logic
 - Added Base64Encode, Base64Decode data udf
 - Added TerminatedSplitN text util function

## Jan 2 2019 - v0.6.2
 - Added FollowRedirects option to http client

## Jan 1 2019 - v0.6.1
 - Patched SortedIterator
 - Patched embedded non pointer struct conversion

## Dec 29 2018 - v0.6.0
 - Added SortedRange, Iterator, SortedIterator to compacted slice
 
## Dec 28 2018 - v0.5.4
 - Added QueryEscape udf
 - Updated handling udf with single quoted literals 

## Dec 27 2018 - v0.5.3
 - Added DecoderFactory method to url.Resource 
 - Patched secret location with URL scheme 

## Dec 26 2018 - v0.5.2
 - Patched KV nested slice conversion 
 - Patched handling unexported fields 
 - Minor patches

## Dec 24 2018 - v0.5.1
 - Patched KV conversion where value was nil
 - Updated secret service location lookup order
 - Minor patches

## Dec 18 2018 - v0.5.0
 - NormalizeKVPairs - to converts slice of KV paris into a map, and map[interface{}]interface{} to map[string]interface{} 
 - Moved stand expandable UDF from neatly project
 - Added data and data/udf documentation

## Dec 7 2018 - v0.4.1
 - Enhanced UDF multi arguments calls
 - Added [] sub map key expression support
 - Patched name with sub references in Map.SetValue

## Dec 7 2018 - v0.4.0
 - Added elapsed/remaining day helper functions: ElapsedDay, ElapsedToday, RemainingToday

## Dec 6 2018 - v0.3.8
 - Patched udf arguments conversion glitch
 - Patched scp service with additional fallback for file scraping
 - Refactor data/map expression parser, Added basic arithmetic support
 - Added expansion of struct datatype, patched asEncodableValue
 
 
## Dec 3 2018 - v0.3.1
 - Refactor data/map expression parser, Added basic arithmetic support
 - Refactor tokenizer matchers 

## Nov 28 2018 - v0.2.4
 - Patched ToInt, ToFloat conversion with nil pointer

## Nov 24 2018 - v0.2.3
 - Added ToBoolean
 - Streamline ssh Session init
     
## Nov 19 2018 - v0.2.2
 - Added error check for opening shell in ssh Session
 - Enhance SSH termination error

## Nov 8 2018 - v0.2.0

 - Added TimeAt utility method for creating dynamically semantic based time.
 - Added IllegalTokenError, ExpectToken and ExpectTokenOptionallyFollowedBy parsing helpers
 - Added RemainingSequenceMatcher
 - Added SSH stdout buffering with listener frequency flush

## Oct 20 2018 - v0.1.1

 - Added Replace method on data/map.go
 - Added path support to Delete method on on data/map

## Jul 1 2016 (Alpha)

  * Initial Release.
