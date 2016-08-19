package system

type set map[*job]struct{}

func (s set) Contains(j *job) (ok bool) {
	_, ok = s[j]
	return
}

func (s set) Put(j *job) {
	s[j] = struct{}{}
}
