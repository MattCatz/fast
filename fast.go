package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ddo/go-fast"
	"github.com/ddo/go-spin"
)

func main() {
	var kb, mb, gb, silent bool
	flag.BoolVar(&kb, "k", false, "Format output in Kbps")
	flag.BoolVar(&mb, "m", false, "Format output in Mbps")
	flag.BoolVar(&gb, "g", false, "Format output in Gbps")
	flag.BoolVar(&silent, "silent", false, "Surpress all output except for the final result")

	flag.Parse()

	if kb && (mb || gb) || (mb && kb) {
		fmt.Println("You may have at most one formating switch. Choose either -k, -m, or -g")
		os.Exit(-1)
	}

	var format func(float64) (string, string, float64)
	if kb {
		format = formatKbps
	} else if mb {
		format = formatMbps
	} else if gb {
		format = formatGbps
	} else {
		format = formatNatural
	}

	spinner := spin.New("")
	ticker := time.NewTicker(100 * time.Millisecond)
	if silent {
		// do no print updates in silent mode
		ticker.Stop()
	}

	// output
	status_update := make(chan string)
	// measure
	KbpsChan := make(chan float64)
	// finish line
	done := make(chan bool)

	printer := func() {
		update := ""
		updates:
		for {
			select {
			case s, ok := <-status_update:
				if ok {
					update = s
				}
			case m, ok := <-KbpsChan:
				if ok {
					f, unit, value := format(m)
					update = fmt.Sprintf(f, value) + " " + unit
				} else {
					// Close up shop
					break updates
				}
			case <- ticker.C:
				// updates get printed with \r
				fmt.Printf("%c[2K %s  %-20s\r", 27, spinner.Spin(), update)
			}
		}


		// final print should use \n
		fmt.Printf("%-20s\n", update)
		done <- true
	}


	go printer()

	fastCom, err := fast.New(&fast.Option{
		BindAddress: "192.168.86.23",
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// init
	err = fastCom.Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	status_update <- "connecting"

	// get urls
	urls, err := fastCom.GetUrls()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	status_update <- "loading"

	err = fastCom.Measure(urls, KbpsChan)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	
	// Wait here till done
	<- done

	return
}

func formatGbps(Kbps float64) (string, string, float64) {
	f := "%.2f"
	unit := "Gbps"
	value := Kbps / 1000000
	return f, unit, value
}

func formatMbps(Kbps float64) (string, string, float64) {
	f := "%.2f"
	unit := "Mbps"
	value := Kbps / 1000
	return f, unit, value
}

func formatKbps(Kbps float64) (string, string, float64) {
	f := "%.f"
	unit := "Kbps"
	value := Kbps
	return f, unit, value
}

func formatNatural(Kbps float64) (string, string, float64) {
	var value float64
	var unit string
	var f string

	if Kbps > 1000000 { // Gbps
		f, unit, value = formatGbps(Kbps)
	} else if Kbps > 1000 { // Mbps
		f, unit, value = formatMbps(Kbps)
	} else {
		f, unit, value = formatKbps(Kbps)
	}
	return f, unit, value
}
