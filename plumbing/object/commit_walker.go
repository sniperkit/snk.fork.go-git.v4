/*
Sniperkit-Bot
- Date: 2018-08-11 15:40:00.935176804 +0200 CEST m=+0.032827986
- Status: analyzed
*/

package object

import (
	"io"

	"github.com/sniperkit/snk.fork.go-git.v4/plumbing"
	"github.com/sniperkit/snk.fork.go-git.v4/plumbing/storer"
)

type commitPreIterator struct {
	seenExternal map[plumbing.Hash]bool
	seen         map[plumbing.Hash]bool
	stack        []CommitIter
	start        *Commit
}

// NewCommitPreorderIter returns a CommitIter that walks the commit history,
// starting at the given commit and visiting its parents in pre-order.
// The given callback will be called for each visited commit. Each commit will
// be visited only once. If the callback returns an error, walking will stop
// and will return the error. Other errors might be returned if the history
// cannot be traversed (e.g. missing objects). Ignore allows to skip some
// commits from being iterated.
func NewCommitPreorderIter(
	c *Commit,
	seenExternal map[plumbing.Hash]bool,
	ignore []plumbing.Hash,
) CommitIter {
	seen := make(map[plumbing.Hash]bool)
	for _, h := range ignore {
		seen[h] = true
	}

	return &commitPreIterator{
		seenExternal: seenExternal,
		seen:         seen,
		stack:        make([]CommitIter, 0),
		start:        c,
	}
}

func (w *commitPreIterator) Next() (*Commit, error) {
	var c *Commit
	for {
		if w.start != nil {
			c = w.start
			w.start = nil
		} else {
			current := len(w.stack) - 1
			if current < 0 {
				return nil, io.EOF
			}

			var err error
			c, err = w.stack[current].Next()
			if err == io.EOF {
				w.stack = w.stack[:current]
				continue
			}

			if err != nil {
				return nil, err
			}
		}

		if w.seen[c.Hash] || w.seenExternal[c.Hash] {
			continue
		}

		w.seen[c.Hash] = true

		if c.NumParents() > 0 {
			w.stack = append(w.stack, filteredParentIter(c, w.seen))
		}

		return c, nil
	}
}

func filteredParentIter(c *Commit, seen map[plumbing.Hash]bool) CommitIter {
	var hashes []plumbing.Hash
	for _, h := range c.ParentHashes {
		if !seen[h] {
			hashes = append(hashes, h)
		}
	}

	return NewCommitIter(c.s,
		storer.NewEncodedObjectLookupIter(c.s, plumbing.CommitObject, hashes),
	)
}

func (w *commitPreIterator) ForEach(cb func(*Commit) error) error {
	for {
		c, err := w.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		err = cb(c)
		if err == storer.ErrStop {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *commitPreIterator) Close() {}

type commitPostIterator struct {
	stack []*Commit
	seen  map[plumbing.Hash]bool
}

// NewCommitPostorderIter returns a CommitIter that walks the commit
// history like WalkCommitHistory but in post-order. This means that after
// walking a merge commit, the merged commit will be walked before the base
// it was merged on. This can be useful if you wish to see the history in
// chronological order. Ignore allows to skip some commits from being iterated.
func NewCommitPostorderIter(c *Commit, ignore []plumbing.Hash) CommitIter {
	seen := make(map[plumbing.Hash]bool)
	for _, h := range ignore {
		seen[h] = true
	}

	return &commitPostIterator{
		stack: []*Commit{c},
		seen:  seen,
	}
}

func (w *commitPostIterator) Next() (*Commit, error) {
	for {
		if len(w.stack) == 0 {
			return nil, io.EOF
		}

		c := w.stack[len(w.stack)-1]
		w.stack = w.stack[:len(w.stack)-1]

		if w.seen[c.Hash] {
			continue
		}

		w.seen[c.Hash] = true

		return c, c.Parents().ForEach(func(p *Commit) error {
			w.stack = append(w.stack, p)
			return nil
		})
	}
}

func (w *commitPostIterator) ForEach(cb func(*Commit) error) error {
	for {
		c, err := w.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		err = cb(c)
		if err == storer.ErrStop {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *commitPostIterator) Close() {}
