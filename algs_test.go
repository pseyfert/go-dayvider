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
	"sort"
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

func dumpblock(b Block, i int) {
	if testing.Verbose() {
		fmt.Printf("Block %d info:\n", i)
		fmt.Printf("  start: %v\n", b.Start)
		fmt.Printf("  end:   %v\n", b.End)
		fmt.Println("Bookings:")
		for bo := b.Seed; bo < b.Last; bo += 1 {
			fmt.Printf("    booking %d\n", bo)
			fmt.Printf("     start: %v\n", b.Event.Bookings[bo].Start)
			fmt.Printf("       end: %v\n", b.Event.Bookings[bo].End)
		}
	}
}

func validateBlock(b Block, i int, t *testing.T) {
	if b.Last <= b.Seed {
		dumpblock(b, i)
		t.Fatal("block does not contain events")
	}
	if b.Seed < 0 || b.Seed >= len(b.Event.Bookings) {
		dumpblock(b, i)
		t.Fatal("block start out of event range")
	}
	if b.Last <= 0 || b.Seed > len(b.Event.Bookings) {
		dumpblock(b, i)
		t.Fatal("block end out of event range")
	}
	if b.Start != b.Event.Bookings[b.Seed].Start {
		dumpblock(b, i)
		t.Fatal("block starts not aligned")
	}

	lastend := b.Event.Bookings[b.Seed].End
	for i := b.Seed; i < b.Last; i += 1 {
		if b.Event.Bookings[i].Start.Before(b.Start) {
			dumpblock(b, i)
			t.Fatal("booking in block start before block")
		}
		if b.Event.Bookings[i].End.After(b.End) {
			dumpblock(b, i)
			t.Fatal("booking in block ends after block")
		}
		if lastend.Before(b.Event.Bookings[i].End) {
			lastend = b.Event.Bookings[i].End
		}
	}
	if lastend != b.End {
		dumpblock(b, i)
		t.Fatal("booking end seems excessive")
	}

	connectedend := b.Event.Bookings[b.Seed].End
	for i := b.Seed; i < b.Last; i += 1 {
		if !b.Event.Bookings[i].Start.After(connectedend) &&
			b.Event.Bookings[i].End.After(connectedend) {
			connectedend = b.Event.Bookings[i].End
			i = b.Seed
		}
	}
	if connectedend.Before(b.End) {
		if testing.Verbose() {
			fmt.Printf("potential gap starting at %v\n", connectedend)
			dumpblock(b, i)
		}
		t.Fatal("gap found within block")
	}
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

		blocks := event.Blockify()
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

		validateBlocks(blocks, t)
		if testing.Verbose() {
			fmt.Println("passed event")
		}
	}
}

func validateBlocks(blocks []Block, t *testing.T) {
	if len(blocks) == 0 {
		t.Fatal("no blocks to validate blocks")
	}

	e := blocks[0].Event

	for i, b := range blocks {
		validateBlock(b, i, t)
		if e != b.Event {
			t.Fatal("blocks from different events unexpected")
		}
	}

	if blocks[0].Seed != 0 {
		t.Fatal("first block doesn't start with first booking")
	}
	if blocks[len(blocks)-1].Last != len(e.Bookings) {
		t.Fatal("last block doesn't end with last booking")
	}

	for i := 1; i < len(blocks); i += 1 {
		if blocks[i-1].Last != blocks[i].Seed {
			t.Fatal("not all bookings seem to have ended up in blocks")
		}
	}
}

func TestTrivialBlocks(t *testing.T) {
Reps:
	for e := 0; e < test_repetitions; e += 1 {
		bookings := make([]Booking, 0, test_bookings)

		dates := make(times, 0, test_bookings*2)
		for b := 0; b < test_bookings; b += 1 {
			dates = append(dates, randate())
			dates = append(dates, randate())
		}
		sort.Sort(dates)
		for b := 0; b < test_bookings; b += 1 {
			bookings = append(bookings, Booking{Start: dates[2*b], End: dates[2*b+1]})
			if b > 0 {
				if bookings[b].Start.Equal(bookings[b-1].End) {
					bookings[b].Start = bookings[b].Start.Add(time.Second)
				}
				if bookings[b].Start.After(bookings[b].End) {
					fmt.Println("have to skip 'TestTrivialBlocks' due to too-close bookings")
					continue Reps
				}
			}
		}
		event := NewEvent(bookings)

		blocks := event.Blockify()

		validateBlocks(blocks, t)

		if len(blocks) != len(bookings) {
			if testing.Verbose() {
				for i, b := range blocks {
					dumpblock(b, i)
				}
			}
			t.Fatalf("should have number of blocks (%d) equal to number of bookings (%d)", len(blocks), len(bookings))
		}
	}
}

func TestLongBlock(t *testing.T) {
	for e := 0; e < test_repetitions; e += 1 {
		bookings := make([]Booking, 0, test_bookings)

		dates := make(times, 0, test_bookings*2)
		for b := 0; b <= test_bookings; b += 1 {
			dates = append(dates, randate())
		}
		sort.Sort(dates)
		for b := 0; b < test_bookings; b += 1 {
			bookings = append(bookings, Booking{Start: dates[b], End: dates[b+1]})
		}
		event := NewEvent(bookings)

		blocks := event.Blockify()

		validateBlocks(blocks, t)

		if len(blocks) != 1 {
			if testing.Verbose() {
				for i, b := range blocks {
					dumpblock(b, i)
				}
			}
			t.Fatalf("should have received only one block rather than %d", len(blocks))
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
