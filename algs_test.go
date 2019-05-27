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
	start := time.Date(2020, 1, 0, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Duration(test_days)*24*time.Hour - time.Minute)
	min := start.Unix()
	max := end.Unix()
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

func dumpwd(wd WrappedDuration) {
	if testing.Verbose() {
		for i := 0; i < len(wd.Starts); i += 1 {
			fmt.Printf("  Duration from %v to %v\n", wd.Starts[i], wd.Ends[i])
		}
	}
}

func validateWrappedDuration(wd WrappedDuration, t *testing.T) {
	if len(wd.Starts) != len(wd.Ends) {
		t.Fatal("inconsistend WrappedDuration")
	}
	if wd.Starts[0] != time.Duration(0) {
		dumpwd(wd)
		t.Fatal("first durations should start at 0")
	}
	for i := 0; i < len(wd.Starts); i += 1 {
		if wd.Starts[i] < time.Duration(0) {
			dumpwd(wd)
			t.Fatal("durations should not start before 0")
		}
		if wd.Ends[i] < wd.Starts[i] {
			dumpwd(wd)
			t.Fatal("duration wrongly ordered")
		}
		if wd.Ends[len(wd.Ends)-1] > 24*time.Hour {
			dumpwd(wd)
			t.Fatal("times wrap later than 24h")
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
Reps:
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
		wd := wrapDurations(blocks)
		validateWrappedDuration(wd, t)

		wblocks := wrapBlocks(blocks)
		validateBlocks(wblocks, t)
		if wblocks[len(wblocks)-1].End.Sub(wblocks[0].Start) > 24*time.Hour {
			t.Fatalf("wrapped blocks span more than 1 day")
		}

		gaps := Gaps(wblocks)

		if test_days == 1 {
			if len(gaps) != len(blocks) {
				t.Fatalf("there should be one gap per block!")
			}
		} else {
			if len(gaps) != len(wblocks) && len(gaps) != len(wblocks)-1 {
				t.Fatalf("there should be more gaps")
			}
		}

		if len(gaps) == 0 {
			if impossibleEvent(wblocks) {
				fmt.Printf("skipping impossible event\n")
				continue Reps
			} else {
				t.Fatalf("no gaps. ensure there are gaps")
			}
		}
		if longestGap(gaps) >= len(gaps) {
			t.Fatalf("found impossible last gap")
		}
		eod, err := EndOfFirstDay(blocks)
		if err != nil {
			t.Fatal("impossible event not caught earlier")
		}
		validateEndOfDay(bookings, eod, t)
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
		if blocks[i].Start.Before(blocks[i-1].Start) {
			t.Fatal("blocks array doesn't seem to be ordered")
		}
		if blocks[i-1].End.After(blocks[i].Start) {
			t.Fatal("overlapping blocks!")
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
		wd := wrapDurations(blocks)
		validateWrappedDuration(wd, t)

		wblocks := wrapBlocks(blocks)
		validateBlocks(wblocks, t)
		if wblocks[len(wblocks)-1].End.Sub(wblocks[0].Start) > 24*time.Hour {
			t.Fatalf("wrapped blocks span more than 1 day")
		}

		gaps := Gaps(wblocks)

		if test_days == 1 {
			if len(gaps) != test_bookings {
				t.Fatalf("there should be one gap per booking!")
			}
		}

		if len(gaps) == 0 {
			if impossibleEvent(wblocks) {
				fmt.Printf("skipping impossible event\n")
				continue Reps
			} else {
				t.Fatalf("no gaps. ensure there are gaps")
			}
		}
		if longestGap(gaps) >= len(gaps) {
			t.Fatalf("found impossible last gap")
		}
		eod, err := EndOfFirstDay(blocks)
		if err != nil {
			t.Fatal("impossible event not caught earlier")
		}
		validateEndOfDay(bookings, eod, t)
	}
}

func TestLongBlock(t *testing.T) {
Reps:
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

		wd := wrapDurations(blocks)
		validateWrappedDuration(wd, t)

		wblocks := wrapBlocks(blocks)
		validateBlocks(wblocks, t)
		if wblocks[len(wblocks)-1].End.Sub(wblocks[0].Start) > 24*time.Hour {
			t.Fatalf("wrapped blocks span more than 1 day")
		}

		if test_days > 1 {
			continue Reps
		}

		gaps := Gaps(wblocks)
		if len(gaps) != 1 {
			t.Fatalf("there must be exactly one gap in the long block test")
		}
		if longestGap(gaps) >= len(gaps) {
			t.Fatalf("found impossible last gap")
		}
		eod, err := EndOfFirstDay(blocks)
		if err != nil {
			t.Fatal("impossible event not caught earlier")
		}
		validateEndOfDay(bookings, eod, t)
	}
}

func validateEndOfDay(bookings []Booking, endofday time.Time, t *testing.T) {
	for _, b := range bookings {
		if endofday.After(b.Start) && endofday.Before(b.End) {
			t.Fatal("found booking during end of day")
		}
	}
	if endofday.Sub(bookings[0].Start) > 24*time.Hour {
		t.Fatal("first day is longer than 24 hours")
	}
}

var test_repetitions int
var test_bookings int
var test_days int

func TestMain(m *testing.M) {
	flag.IntVar(&test_repetitions, "repetitions", 5, "repeat randomized test N times")
	flag.IntVar(&test_bookings, "bookings", 3, "test with a random event of N bookings")
	flag.IntVar(&test_days, "days", 3, "test all bookings spread over N days")
	flag.Parse()
	// call flag.Parse() here if TestMain uses flags
	os.Exit(m.Run())
}
