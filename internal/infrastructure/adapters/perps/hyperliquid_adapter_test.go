package perps

import "testing"

func TestParseL2BookPrice(t *testing.T) {
	message := []byte(`{
		"channel":"l2Book",
		"data":{
			"levels":[
				[{"px":"2055.25","sz":"1.2"}],
				[{"px":"2055.75","sz":"0.8"}]
			]
		}
	}`)

	price, ok := parseL2BookPrice(message)
	if !ok {
		t.Fatalf("expected parser to extract price")
	}
	if price != 2055.25 {
		t.Fatalf("expected 2055.25, got %f", price)
	}
}

func TestParseUserEventPositionSizeFromFill(t *testing.T) {
	message := []byte(`{
		"channel":"userEvents",
		"data":{
			"fills":[
				{"sz":"0.15"},
				{"sz":"0.30"}
			]
		}
	}`)

	size, ok := parseUserEventPositionSize(message)
	if !ok {
		t.Fatalf("expected parser to extract size")
	}
	if size != 0.30 {
		t.Fatalf("expected 0.30, got %f", size)
	}
}

func TestParseUserEventPositionSizeFromClearinghouseState(t *testing.T) {
	message := []byte(`{
		"channel":"userEvents",
		"data":{
			"clearinghouseState":{
				"assetPositions":[
					{"position":{"coin":"ETH","szi":"-0.42"}}
				]
			}
		}
	}`)

	size, ok := parseUserEventPositionSize(message)
	if !ok {
		t.Fatalf("expected parser to extract size")
	}
	if size != 0.42 {
		t.Fatalf("expected 0.42, got %f", size)
	}
}
