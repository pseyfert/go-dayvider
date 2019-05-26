/*
 * Copyright (C) 2019 Paul Seyfert
 * Author: Paul Seyfert <pseyfert@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package dayvider

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

func randombooking() Booking {
	a := randate()
	b := randate()

	if a.Before(b) {
		return Booking{Start: a, End: b}
	}
	return Booking{Start: b, End: a}
}

func randate() time.Time {
	// modified from https://stackoverflow.com/a/43497333
	min := time.Date(2000, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	max := time.Date(2020, 9, 9, 9, 9, 9, 9, time.UTC).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}

func dumpblock(b Block) {
	if testing.Verbose() {
		fmt.Println("Block info:")
		fmt.Printf("  start: %v\n", b.Start)
		fmt.Printf("  end:   %v\n", b.End)
		fmt.Println("Bookings:")
		for bo := b.Seed; bo < b.Last; bo += 1 {
			fmt.Printf("     start: %v\n", b.Event.Bookings[bo].Start)
			fmt.Printf("       end: %v\n", b.Event.Bookings[bo].End)
		}
	}
}

func validateBlock(b Block, t *testing.T) {
	if b.Last <= b.Seed {
		dumpblock(b)
		t.Fatal("block does not contain events")
	}
	if b.Seed < 0 || b.Seed >= len(b.Event.Bookings) {
		dumpblock(b)
		t.Fatal("block start out of event range")
	}
	if b.Last <= 0 || b.Seed > len(b.Event.Bookings) {
		dumpblock(b)
		t.Fatal("block end out of event range")
	}
	if b.Start != b.Event.Bookings[b.Seed].Start {
		dumpblock(b)
		t.Fatal("block starts not aligned")
	}

	lastend := b.Event.Bookings[0].End
	for i := b.Seed; i < b.Last; i += 1 {
		if b.Event.Bookings[i].Start.Before(b.Start) {
			dumpblock(b)
			t.Fatal("booking in block start before block")
		}
		if b.Event.Bookings[i].End.After(b.End) {
			dumpblock(b)
			t.Fatal("booking in block ends after block")
		}
		if lastend.Before(b.Event.Bookings[i].End) {
			lastend = b.Event.Bookings[i].End
		}
	}
	if lastend != b.End {
		dumpblock(b)
		t.Fatal("booking end seems excessive")
	}
	// todo: validate absence of gaps in a block
}

func TestBlocking(t *testing.T) {
	for e := 0; e < test_repetitions; e += 1 {
		bookings := make([]Booking, 0, test_bookings)
		for b := 0; b < test_bookings; b += 1 {
			bookings = append(bookings, randombooking())
		}
		if testing.Verbose() {
			fmt.Println("generated bookings")
		}
		event := NewEvent(bookings)
		if testing.Verbose() {
			fmt.Println("generated event")
		}

		blocks := event.blockify()
		if testing.Verbose() {
			fmt.Println("blockified")
		}

		if testing.Verbose() {
			fmt.Printf("generated event with %d bookings\n", len(event.Bookings))
			fmt.Printf("generated %d blocks\n", len(blocks))
		}

		if len(blocks) == 0 {
			t.Fatal("no blocks for an event")
		}

		for _, block := range blocks {
			validateBlock(block, t)
			if testing.Verbose() {
				fmt.Println("block passed")
			}
		}
		if testing.Verbose() {
			fmt.Println("passed event")
		}
	}

}

var test_repetitions int
var test_bookings int

func TestMain(m *testing.M) {
	flag.IntVar(&test_repetitions, "repetitions", 5, "repeat randomized test N times")
	flag.IntVar(&test_bookings, "bookings", 3, "test with a random event of N bookings")
	flag.Parse()
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}
