package jd_cookie

import "github.com/cdle/sillyGirl/core"

var sillyGirl = core.NewBucket("sillyGirl")
var name = "芝士"

func init() {
	if sillyGirl.Get("name") != name {
		sillyGirl.Set("name", "芝士")
	}
}
