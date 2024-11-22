package utils

import (
	"testing"

	"github.com/slack-go/slack"
)

func TestShuffle(t *testing.T) {
	arr := []string{"a", "b", "c", "d", "e"}
	Shuffle(arr)

	// check arr was shuffled
	if arr[0] == "a" && arr[1] == "b" && arr[2] == "c" && arr[3] == "d" && arr[4] == "e" {
		t.Errorf("Array was not shuffled")
	}
}

func TestCanAddToList(t *testing.T) {
	var TSQuery = "T&S"
	var PlatformQuery = "Platform"
	var ProductQuery = "Product"
	var FEQuery = "FE"
	var BEQuery = "BE"

	info := slack.User{
		Profile: slack.UserProfile{
			Title: "[T&S] [FE]",
		},
	}

	// T&S queries
	if !CanAddToList(&info, TSQuery, true) {
		t.Errorf("User should be added to list")
	}
	info.Profile.Title = "T&S [BE]"
	if CanAddToList(&info, TSQuery, true) {
		t.Errorf("User should not be added to list")
	}

	// Platform queries
	info.Profile.Title = "[Platform] [BE]"
	if !CanAddToList(&info, PlatformQuery, true) {
		t.Errorf("User should be added to list")
	}
	if !CanAddToList(&info, BEQuery, true) {
		t.Errorf("User should be added to list")
	}
	info.Profile.Title = "Platform [FE]"
	if CanAddToList(&info, PlatformQuery, true) {
		t.Errorf("User should not be added to list")
	}

	// Product queries
	info.Profile.Title = "[Product] [BE]"
	if !CanAddToList(&info, ProductQuery, true) {
		t.Errorf("User should be added to list")
	}
	if !CanAddToList(&info, BEQuery, true) {
		t.Errorf("User should be added to list")
	}
	info.Profile.Title = "Product [FE]"
	if CanAddToList(&info, ProductQuery, true) {
		t.Errorf("User should not be added to list")
	}
	if !CanAddToList(&info, FEQuery, true) {
		t.Errorf("User should be added to list")
	}
}
