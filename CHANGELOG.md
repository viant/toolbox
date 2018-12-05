## Dec 5 2018 - v0.3.4
 - Patched scp service with additional fallback for file scraping
 - Refactor data/map expression parser, added basic arithmetic support
 - Added expansion of struct datatype, patched asEncodableValue
 
 
## Dec 3 2018 - v0.3.1
 - Refactor data/map expression parser, added basic arithmetic support
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
