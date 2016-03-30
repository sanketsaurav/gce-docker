package manager

import (
	"time"

	. "gopkg.in/check.v1"
)

type WorkerSuite struct{}

var _ = Suite(&WorkerSuite{})

func (s *WorkerSuite) TestAdd(c *C) {
	delay := 10 * time.Millisecond
	start := time.Now()
	var since time.Duration

	w := NewWorker()
	w.Add(JobID(""), func() error {
		since = time.Since(start)
		return nil
	}, delay)

	time.Sleep(delay * 2)
	c.Assert(since > delay, Equals, true)
}

func (s *WorkerSuite) TestAddAndDelete(c *C) {
	id := JobID("foo")
	delay := 10 * time.Millisecond
	start := time.Now()
	var since time.Duration

	w := NewWorker()
	w.Add(id, func() error {
		since = time.Since(start)
		return nil
	}, delay)

	w.Delete(id)

	time.Sleep(delay * 2)

	c.Assert(since, Equals, time.Duration(0))
}
