package synchronizer

type Statistics struct {
	TotalFiles     uint64
	UploadedFiles  uint64
	ProcessedFiles uint64
	shouldNotify   bool
	statNotify     chan Statistics
}

func newStats(totalFiles uint64, statNotify chan Statistics, shouldNotify bool) *Statistics {
	stats := &Statistics{TotalFiles: totalFiles, statNotify: statNotify, shouldNotify: shouldNotify}
	stats.notify()
	return stats
}

func (s *Statistics) Uploaded() {
	s.UploadedFiles++
	s.notify()
}

func (s *Statistics) Processed() {
	s.ProcessedFiles++
	s.notify()
}

func (s *Statistics) Progress() float32 {
	if s.TotalFiles == 0 {
		return float32(0)
	}
	return float32(s.ProcessedFiles) / float32(s.TotalFiles)
}

func (s *Statistics) notify() {
	if s.shouldNotify {
		s.statNotify <- *s
	}
}
