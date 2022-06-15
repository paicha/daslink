package example

import (
	"fmt"
	"regexp"
	"testing"
)

func TestLink(t *testing.T) {
	//value := "ipfs://PAIg0cuUJFfX7pBunqLLgQIaeTt8tEJtBIrPSzuxWkD3BA"
	//value = "PAIg0cuUJFfX7pBunqLLgQIaeTt8tEJtBIrPSzuxWkD3BA"
	//value = "https://siasky.net/PAIg0cuUJFfX7pBunqLLgQIaeTt8tEJtBIrPSzuxWkD3BA"
	//value = "https://siasky.net/XAELj697T77MQOhB2XD-GGNY0bSel0xla9mS1p7L50bv6w"
	value := "https://siasky.net/AAAEn_sRJSoC8N95In4b9E-n23C9udgTO-NkrrKGN4QDNg"
	fmt.Println(len("AAAEn_sRJSoC8N95In4b9E-n23C9udgTO-NkrrKGN4QDNg"))
	re := regexp.MustCompile(`([0-9A-Za-z-_]{46})`)
	if results := re.FindStringSubmatch(value); len(results) == 2 {
		value = results[1]
	}
	fmt.Println(value)
}
