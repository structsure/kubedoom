package main

import "log"

func debugSlice(mySlice []string) {
	debugSliceAt(mySlice, "Debug")
}
func debugSliceAt(mySlice []string, debugMsg string) {
	log.Printf("\n+++++ %v +++++\n", debugMsg)
	printSlice(mySlice)
	log.Printf("\n----- %v -----\n", debugMsg)
}
func printSlice(s []string) {
	for _, v := range s {
		log.Print(v)
	}
}
