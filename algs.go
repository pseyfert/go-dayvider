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

import "time"

func getblock(e *Event, seed int) (retval Block) {
	retval.Event = e
	retval.Seed = seed
	retval.Start = e.Bookings[seed].Start
	retval.End = e.Bookings[seed].End
	retval.Last = retval.Seed + 1

	for i := retval.Seed; i < len(e.Bookings); i += 1 {

		// further bookings might:
		//  * have ended before the current `end`   -> skip
		//  * start after the current `end`         -> break
		//  * be ongoing                            -> update `end`

		if e.Bookings[i].Start.After(retval.End) {
			retval.Last = i // upper bound is the first booking that's not part of the current block
			break
		} else {
			retval.Last = i + 1 // push upper bound further, just in case
		}
		if e.Bookings[i].End.Before(retval.End) {
			continue
		}
		retval.End = e.Bookings[i].End
	}
	return
}

func (e *Event) Blockify() []Block {
	blocks := make([]Block, 0)

	seed := 0

	for {
		if seed == len(e.Bookings) {
			break
		}

		newblock := getblock(e, seed)
		seed = newblock.Last
		blocks = append(blocks, newblock)
	}

	return blocks
}

func WrapDurations(blocks []Block) WrappedDuration {
	var retval WrappedDuration
	retval.Ref = blocks[0].Start
	for _, b := range blocks {
		retval.Starts = append(retval.Starts, b.Start.Sub(retval.Ref)%(time.Duration(24)*time.Hour))
		retval.Ends = append(retval.Ends, b.End.Sub(retval.Ref)%(time.Duration(24)*time.Hour))

		if b.End.Sub(b.Start) > 24*time.Hour {
			var cornercase WrappedDuration
			cornercase.Ref = blocks[0].Start
			cornercase.Starts = []time.Duration{0}
			cornercase.Ends = []time.Duration{24 * time.Hour}
			return cornercase
		}

		i := len(retval.Starts) - 1
		if retval.Starts[i] > retval.Ends[i] {
			tmp := retval.Ends[i]
			retval.Ends[i] = time.Duration(24) * time.Hour
			retval.Starts = append(retval.Starts, time.Duration(0))
			retval.Ends = append(retval.Ends, tmp)
		}
	}

	return retval
}

func DurationsToBookings(wd WrappedDuration) (retval []Booking) {
	retval = make([]Booking, 0, len(wd.Starts))

	for i := 0; i < len(wd.Starts); i += 1 {
		retval = append(retval, Booking{Start: wd.Ref.Add(wd.Starts[i]), End: wd.Ref.Add(wd.Ends[i])})
	}

	return
}

func Gaps(durationbookings []Block) (retval []time.Duration) {
	retval = make([]time.Duration, 0, len(durationbookings))

	for i := 0; i < len(durationbookings)-1; i += 1 {
		end := durationbookings[i+1].Start
		start := durationbookings[i].End
		retval = append(retval, end.Sub(start))
	}
	i := len(durationbookings) - 1
	end := durationbookings[0].Start.Add(24 * time.Hour)
	if start := durationbookings[i].End; start.Before(end) {
		retval = append(retval, end.Sub(start))
	}

	return
}

func WrapBlocks(blocks []Block) []Block {
	wd := WrapDurations(blocks)
	wb := DurationsToBookings(wd)
	event := NewEvent(wb)
	return event.Blockify()
}

func longestGap(gaps []time.Duration) int {
	longestplace := 0
	longesttime := gaps[0]
	for i, d := range gaps {
		if d > longesttime {
			longesttime = d
			longestplace = i
		}
	}
	return longestplace
}

func EndOfFirstDay(blocks []Block) time.Time {
	wrappedblocks := WrapBlocks(blocks)
	gaps := Gaps(wrappedblocks)
	longestgap := longestGap(gaps)

	return wrappedblocks[longestgap].End
}
