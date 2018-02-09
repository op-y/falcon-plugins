/*
* re.go - functions related to text processing
*
* history
* --------------------
* 2018/1/12, by Ye Zhiqin, create
*
 */

package main

import (
	"strings"
)

/*
* MatchRequest - match the request uri
*
* PARAMS:
*   - request: request in log
*   - pattern: certain uri
*
* RETURNS:
*   - true: if match
*   - false: if not match
 */
func MatchRequest(request, pattern string) bool {
	ok := strings.Contains(request, pattern)
	return ok
}

/*
* GetID - get shop ID in request body
*
* PARAMS:
*   - body: request body in log
*
* RETURNS:
*   - true, id, nil: if found
*   - false, "": if not found
 */
func GetID(body string) (bool, string) {
	fields := strings.Split(body, "&")
	for _, field := range fields {
		kv := strings.Split(field, "=")
		if len(kv) != 2 {
			continue
		}
		k := kv[0]
		if k != "shopID" {
			continue
		}
		v := kv[1]
		return true, v
	}
	return false, ""
}
