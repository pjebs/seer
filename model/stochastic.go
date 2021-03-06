/*
 * Copyright (C) 2018 The Seer Authors. All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package model

import (
	"math"

	"github.com/cshenton/seer/dist/mv"
	"github.com/cshenton/seer/dist/uv"
	"github.com/cshenton/seer/kalman"
	"gonum.org/v1/gonum/mat"
)

// Stochastic is the type against which we apply stochastic model updates.
type Stochastic struct {
	*mv.Normal
}

// NewStochastic returns a stochastic with a proper state prior.
func NewStochastic() (s *Stochastic) {
	loc := []float64{0}
	cov := []float64{1e12}
	n, _ := mv.NewNormal(loc, cov)
	s = &Stochastic{n}
	return s
}

// State returns the kalman filter State.
func (s *Stochastic) State() (k *kalman.State) {
	l := mat.NewDense(s.Dim(), 1, s.Location)
	c := mat.NewDense(s.Dim(), s.Dim(), s.Covariance)

	k, _ = kalman.NewState(l, c)
	return k
}

// System generates process and observation matrices for this linear system.
func (s *Stochastic) System(noise, walk float64) (k *kalman.System) {
	a := mat.NewDense(1, 1, []float64{1})
	b := mat.NewDense(1, 1, []float64{1})
	c := mat.NewDense(1, 1, []float64{1})
	q := mat.NewDense(1, 1, []float64{walk})
	r := mat.NewDense(1, 1, []float64{noise})

	k, _ = kalman.NewSystem(a, b, c, q, r)
	return k
}

// Update performs a filter step against the stochastic state.
func (s *Stochastic) Update(noise, walk, val float64) (err error) {
	st := s.State()
	sy := s.System(noise, walk)

	statePred, _ := kalman.Predict(st, sy)
	newState, _, _ := kalman.Update(statePred, sy, val)

	s.Location = DenseValues(newState.Loc)
	s.Covariance = DenseValues(newState.Cov)
	return nil
}

// Forecast returns a forecasted slice of normal RVs for this stochastic component.
func (s *Stochastic) Forecast(noise, walk float64, n int) (f []*uv.Normal) {
	f = make([]*uv.Normal, n)

	st := s.State()
	sy := s.System(noise, walk)

	for i := 0; i < n; i++ {
		st, _ = kalman.Predict(st, sy)
		ob, _ := kalman.Observe(st, sy)
		f[i] = &uv.Normal{
			Location: DenseValues(ob.Loc)[0],
			Scale:    math.Sqrt(DenseValues(ob.Cov)[0]),
		}
	}
	return f
}
