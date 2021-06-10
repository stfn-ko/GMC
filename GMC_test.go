package main

import (
	"fmt"
	"testing"
	"time"
)

// TEST CACHE (generic memory cache)
var test_cache GenericMemoryCache = &CACHE{}

/* = FORMATTING = */

// This struct contains formatting elements for testing funcs. You can esily add new string fields to improve your testing expirience
type format struct {
	info_on_timing_didnt_exp,
	err_on_timing_didnt_exp,
	separation_line_long,
	info_on_timing_exp,
	confirmation_line,
	err_on_timing_exp,
	separation_line,
	info_on_req,
	info_on_del,
	test_data,
	empty string
}

// This func initializes fields of -format- struct. You can esily add new string elements to improve your testing expirience
func get() *format {
	new_format := &format{
		info_on_timing_didnt_exp: "%vSuccess! %v was found!\nData inside: %v\nExpiration time set: %v\tTime gone since expiration time was set: %v\n%v",
		err_on_timing_didnt_exp:  "%vFound %v -> should have had already expired\nExpiration time set: %v\tTime gone since expiration time was set: %v\n%v",
		separation_line_long:	  "-----------------------------------------------------------------------\n",
		info_on_timing_exp:       "%vSuccess! %v has expired on time!\nData inside: %v\nExpiration time set: %v\tTime gone since expiration time was set: %v\n%v",
		err_on_timing_exp:        "%vCouldn't find %v -> shouldn't have had expired yet\nExpiration time set: %v\tTime gone since expiration time was set: %v\n%v",
		confirmation_line:        "%vCache line has been created succsessfully!\n%v",
		separation_line:          "------------------------------------------------------------\n",
		info_on_req:              "%vRequested entry with key '%v' was found: %t\nData contained in this entry: %v\n%v",
		info_on_del:              "%vRequested entry with key '%v' has been deleted: %t\nData contained in this entry: %v\n%v",
		test_data:                "some data",
		empty:                    "",
	}
	return new_format
}

/* = UNIT TESTS = */ // *Check out -format- struct and -get- func in =FORMATTING= for more info

func Test_Cache(t *testing.T) {

	test_cache = New(DefaultExpiration, 1*time.Millisecond)
	fmt.Printf(get().confirmation_line, get().separation_line, get().separation_line)
}

func Test_Get_Set(t *testing.T) {

	data, found := test_cache.Get("a")
	if found {
		t.Errorf("'a' was found before it was created.")
	} else {
		fmt.Printf(get().info_on_req, get().separation_line, "a", found, data, get().separation_line)
	}
	// {
	test_cache.Set("a", get().test_data, NoExpiration)
	test_cache.Set("b", get().test_data, 2*time.Millisecond)
	// }

	data1, found1 := test_cache.Get("a")
	if !found1 {
		t.Errorf("'a' was not found after it got set to NoExpiration.")
	} else {
		fmt.Printf(get().info_on_req, get().empty, "a", found1, data1, get().separation_line)
	}
	<-time.After(3 * time.Millisecond) // {			// Testing Cache times

	data2, found2 := test_cache.Get("a")
	if !found2 {
		t.Errorf("'a' was not found due to expiration (a was set to NoExpiration setting).")
	} else {
		fmt.Printf(get().info_on_req, get().empty, "a", found2, data2, get().separation_line)
	}

	data3, found3 := test_cache.Get("b")
	if found3 {
		t.Errorf("'b' was found, but it should have got expired and deleted.")
	} else {
		fmt.Printf(get().info_on_req, get().empty, "b", found3, data3, get().separation_line)
	}
	// }
}

func Test_Delete_Cache(t *testing.T) {

	test_cache.Delete("a") //NoExpiration entries can be deleted only manually

	data1, found := test_cache.Get("a")

	if found {
		t.Error("'a' was found, but it should have been deleted")

	} else if data1 != nil {
		t.Error("data is not nil: ", data1)
	} else {
		fmt.Printf(get().info_on_del, get().separation_line, "a", !found, data1, get().separation_line)
	}
}

/*

 */

func Test_AutoClean(t *testing.T) {

	test_cache = New(DefaultExpiration, 1*time.Millisecond)

	
	test_cache.Set("a", get().test_data, 250*time.Millisecond)
	test_cache.Set("b", get().test_data, 100*time.Millisecond)
	test_cache.Set("c", get().test_data, 110*time.Millisecond)
	test_cache.Set("d", get().test_data, 40*time.Millisecond)
	test_cache.Set("e", get().test_data, 160*time.Millisecond)
	test_cache.Set("f", get().test_data, NoExpiration)
	
	start := time.Now()
	
	if data_b, found_b := test_cache.Get("b"); !found_b { //100ms
		t.Errorf(get().err_on_timing_exp, get().separation_line_long, "b", "100ms", time.Since(start).Milliseconds(), get().separation_line_long)
	} else {
		fmt.Printf(get().info_on_timing_didnt_exp, get().separation_line_long, "b", data_b, "100ms", time.Since(start).Milliseconds(), get().separation_line_long)
	}

	if data_d, found_d := test_cache.Get("d"); !found_d { //40ms
		t.Errorf(get().err_on_timing_exp, get().empty, "d", "40ms", time.Since(start).Milliseconds(), get().separation_line_long)
	} else {
		fmt.Printf(get().info_on_timing_didnt_exp, get().empty, "d", data_d, "40ms", time.Since(start).Milliseconds(), get().separation_line_long)
	}

<-time.After(40 * time.Millisecond) //{

	if data_b, found_b := test_cache.Get("b"); !found_b { //100ms
		t.Errorf(get().err_on_timing_exp, get().empty, "b", "100ms", time.Since(start).Milliseconds(), get().separation_line_long)
	} else {
		fmt.Printf(get().info_on_timing_didnt_exp, get().empty, "b", data_b, "100ms", time.Since(start).Milliseconds(), get().separation_line_long)
	}

	if data_c, found_c := test_cache.Get("c"); !found_c { //110ms
		t.Errorf(get().err_on_timing_exp, get().empty, "c", "110ms", time.Since(start).Milliseconds(), get().separation_line_long)
	} else {
		fmt.Printf(get().info_on_timing_didnt_exp, get().empty, "c", data_c, "110ms", time.Since(start).Milliseconds(), get().separation_line_long)
	}

	if data_d, found_d := test_cache.Get("d"); found_d { //40ms
		t.Errorf(get().err_on_timing_didnt_exp, get().empty, "d", "40ms", time.Since(start).Milliseconds(), get().separation_line_long)
	} else {
		fmt.Printf(get().info_on_timing_exp, get().empty, "d", data_d, "40ms", time.Since(start).Milliseconds(), get().separation_line_long)
	}
	//}

<-time.After(90 * time.Millisecond) //{

	if data_e, found_e := test_cache.Get("e"); !found_e { //160ms
		t.Errorf(get().err_on_timing_exp, get().empty, "e", "160ms", time.Since(start).Milliseconds(), get().separation_line_long)
	} else {
		fmt.Printf(get().info_on_timing_didnt_exp, get().empty, "e", data_e, "160ms", time.Since(start).Milliseconds(), get().separation_line_long)
	}

	if data_a, found_a := test_cache.Get("a"); !found_a { //250ms
		t.Errorf(get().err_on_timing_exp, get().empty, "a", "250ms", time.Since(start).Milliseconds(), get().separation_line_long)
	} else {
		fmt.Printf(get().info_on_timing_didnt_exp, get().empty, "a", data_a, "250ms", time.Since(start).Milliseconds(), get().separation_line_long)
	}
//}
	
<-time.After(200 * time.Millisecond) //{

	if data_a, found_a := test_cache.Get("a"); found_a { //250ms
		t.Errorf(get().err_on_timing_didnt_exp, get().empty, "a", "250ms", time.Since(start).Milliseconds(), get().separation_line_long)
	} else {
		fmt.Printf(get().info_on_timing_exp, get().empty, "a", data_a, "250ms", time.Since(start).Milliseconds(), get().separation_line_long)
	}
//}

<-time.After(5 * time.Second) //{

	if data_f, found_f := test_cache.Get("f"); !found_f { //DefaultExpiration
		t.Errorf(get().err_on_timing_exp, get().empty, "f", "DefaultExpiration", time.Since(start).Milliseconds(), get().separation_line_long)
	} else {
		fmt.Printf(get().info_on_timing_didnt_exp, get().empty, "f", data_f, "DefaultExpiration", time.Since(start).Milliseconds(), get().separation_line_long)
	}
}

