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

func getblock(e *Event, seed int) (retval Block) {
	retval.Event = e
	retval.Seed = seed
	retval.Start = e.Bookings[seed].Start
	retval.End = e.Bookings[seed].End
	retval.Last = retval.Seed + 1

	for i := range e.Bookings[retval.Seed:len(e.Bookings)] {

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

func (e *Event) blockify() []Block {
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
