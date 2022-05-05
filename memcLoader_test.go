package main

import (
	"testing"
)


func TestCahingCorrectTrack(t *testing.T) {
    key := "dvid"
	uuid := "8c4a88e21e2753663561b6f26792ef27"
	lat := "169.423781883"
	lon := "13.1686319629"
	tail := []string{"324324", "34324"}
	expectedTrack := Tracker{
		Key: key,
		Uuid: uuid,
		Lat: lat,
		Lon: lon,
		Tail: tail,
	}
    expectedGoodCounter, expectedErrCounter := 1, 0
	stringTrack := "dvid 8c4a88e21e2753663561b6f26792ef27 169.423781883 13.1686319629 324324 34324"
	bytesTrack := []byte(stringTrack)
	tracks, goodCounter, errCounter := parseBuff(bytesTrack)

	if goodCounter != expectedGoodCounter {
		t.Fatalf("Want %v\n, but got %v", expectedGoodCounter, goodCounter)
	}
	if errCounter != expectedErrCounter {
		t.Fatalf("Want %v\n, but got %v", expectedErrCounter, errCounter)
	}
    for _, track := range tracks {
    	if track.Key != expectedTrack.Key {
	    	t.Fatalf("Want %v\n, but got %v", expectedTrack.Key, track.Key)
     	}
		if track.Uuid != expectedTrack.Uuid {
	    	t.Fatalf("Want %v\n, but got %v", expectedTrack.Uuid, track.Uuid)
     	}
		if track.Lat != expectedTrack.Lat {
	    	t.Fatalf("Want %v\n, but got %v", expectedTrack.Lat, track.Lat)
     	}
		if track.Lon != expectedTrack.Lon {
	    	t.Fatalf("Want %v\n, but got %v", expectedTrack.Lon, track.Lon)
     	}
	}
}