package anys2s

import (
	"sort"

	"github.com/unixpickle/anynet/anysgd"
	"github.com/unixpickle/anyvec"
)

// A Sample is a training sequence with a corresponding
// desired output sequence.
type Sample struct {
	Input  []anyvec.Vector
	Output []anyvec.Vector
}

// A SampleList is an anysgd.SampleList that produces
// sequence-to-sequence samples.
type SampleList interface {
	anysgd.SampleList

	GetSample(idx int) *Sample
}

// A SliceSampleList is a concrete SampleList with
// predetermined samples.
type SliceSampleList []*Sample

// Len returns the number of samples.
func (s SliceSampleList) Len() int {
	return len(s)
}

// Swap swaps two samples.
func (s SliceSampleList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Slice copies a sub-slice of the list.
func (s SliceSampleList) Slice(i, j int) anysgd.SampleList {
	return append(SliceSampleList{}, s[i:j]...)
}

// GetSample returns the sample at the index.
func (s SliceSampleList) GetSample(idx int) *Sample {
	return s[idx]
}

// A SortSampleList wraps a SampleList and ensures that
// samples will be sorted within reasonably small chunks.
// This is often beneficial for RNNs on a GPU, since it
// helps to keep batch sizes stable across timesteps.
type SortSampleList struct {
	SampleList

	// BatchSize is the size of the chunks that should be
	// sorted.
	BatchSize int
}

// Slice produces a subset of the SortSampleList.
func (s *SortSampleList) Slice(i, j int) anysgd.SampleList {
	return &SortSampleList{
		SampleList: s.SampleList.Slice(i, j).(SampleList),
		BatchSize:  s.BatchSize,
	}
}

// PostShuffle sorts batches of sequences.
func (s *SortSampleList) PostShuffle() {
	for i := 0; i < s.Len(); i += s.BatchSize {
		bs := s.BatchSize
		if bs > s.Len()-i {
			bs = s.Len() - i
		}
		s := &sorter{S: s.SampleList, Start: i, End: i + bs}
		sort.Sort(s)
	}
}

type sorter struct {
	S     SampleList
	Start int
	End   int
}

func (s *sorter) Len() int {
	return s.End - s.Start
}

func (s *sorter) Swap(i, j int) {
	s.S.Swap(i+s.Start, j+s.Start)
}

func (s *sorter) Less(i, j int) bool {
	return len(s.S.GetSample(i+s.Start).Input) < len(s.S.GetSample(j+s.Start).Input)
}
