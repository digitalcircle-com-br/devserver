package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	for _, i := range []int{
		16,
		32,
		64,
		128,
		256,
		512,
	} {
		istr := fmt.Sprintf("%d", i)
		i2str := fmt.Sprintf("%d", i*2)
		cmd := exec.Command("sips", "-z", istr, istr, "logo.png", "--out", fmt.Sprintf("lib.iconset/icon_%dx%d.png", i, i))
		bs, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Error: %s", err.Error())
		}
		log.Printf("Output: %s", string(bs))

		cmd = exec.Command("sips", "-z", i2str, i2str, "logo.png", "--out", fmt.Sprintf("lib.iconset/icon_%dx%d@2x.png", i, i))
		bs, err = cmd.CombinedOutput()
		if err != nil {
			log.Printf("Error: %s", err.Error())
		}
		log.Printf("Output: %s", string(bs))
	}

	cmd := exec.Command("iconutil", "-c", "icns", "-o", "icon.icns", "lib.iconset")
	bs, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
	log.Printf("Output: %s", string(bs))
}
