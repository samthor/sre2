
package main

/** Ordered bitset. */
type obitset struct {
  bwords int
  bits []int
  result []int
  pos int
}

/**
 * Create a new ordered bitset. States is the maximum state # that may be saved.
 * Size is the maximum number of states that may be saved.
 */
func NewOBitSet(states int, size int) *obitset {
  bwords := (states+32)>>5 // TODO: we just use lower 32 bits, even if int is int64
  return &obitset{bwords, make([]int, bwords), make([]int, size), 0}
}

/** Puts the given value into the obitset. Returns false if success, true if already exists. */
func (o *obitset) Put(v int) bool {
  shift := v & 31

  // grah can't convert int to byte easily *hatehatehate*
  var value byte
  for j := 0; j < shift; j++ {
    value += 1
  }

  if (o.bits[v>>5] & (1 << value)) == 0 {
    o.bits[v>>5] |= (1 << value)
    o.result[o.pos] = v
    o.pos += 1
    return false
  }
  return true
}

/** Clear obitset. */
func (o *obitset) Clear() {
  for i := 0; i < o.bwords; i++ {
    o.bits[i] = 0
  }
  o.pos = 0
}

/** Get contents of obitset as slice. */
func (o *obitset) Contents() []int {
  return o.result[0:o.pos]
}

/** Length of results. */
func (o *obitset) Length() int {
  return o.pos
}
