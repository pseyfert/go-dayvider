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
	"sort"
	"time"
)

type Event struct {
	Bookings []Booking // assumed sorted by start time
}

type times []time.Time

func (e Event) Less(i, j int) bool {
	return e.Bookings[i].Start.Before(e.Bookings[j].Start)
}

func (e Event) Sort() {
	sort.Sort(e)
}

func (e Event) Len() int {
	return len(e.Bookings)
}

func (e Event) Swap(i, j int) {
	e.Bookings[i], e.Bookings[j] = e.Bookings[j], e.Bookings[i]
}

func (e times) Less(i, j int) bool {
	return e[i].Before(e[j])
}

func (e times) Sort() {
	sort.Sort(e)
}

func (e times) Len() int {
	return len(e)
}

func (e times) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func NewEvent(bs []Booking) (e Event) {
	e.Bookings = make([]Booking, 0, len(bs))
	for _, b := range bs {
		e.Bookings = append(e.Bookings, b)
	}
	e.Sort()
	return
}

type Booking struct {
	Start time.Time
	End   time.Time
}

type Block struct {
	Event *Event
	Seed  int
	Last  int
	Start time.Time
	End   time.Time
}
