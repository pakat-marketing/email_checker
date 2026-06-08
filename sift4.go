package emailchecker

// DistanceFunc returns the edit distance between two strings. A lower value
// means the strings are more similar. It is used to score candidate domains
// when suggesting corrections. The default is sift4.
type DistanceFunc func(a, b string) int

// sift4MaxOffset is the search window sift4 uses to recover from a mismatch.
// It matches the default used by mailcheck.js.
const sift4MaxOffset = 5

// sift4 is a port of the sift4 string distance used by mailcheck.js. It
// operates on bytes, which is correct for the ASCII domain names it scores.
func sift4(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	l1 := len(s1)
	l2 := len(s2)

	c1, c2 := 0, 0
	lcss := 0
	localCS := 0
	trans := 0

	type offset struct {
		c1, c2 int
		trans  bool
	}
	var offsetArr []offset

	for c1 < l1 && c2 < l2 {
		if s1[c1] == s2[c2] {
			localCS++
			isTrans := false
			i := 0
			for i < len(offsetArr) {
				ofs := &offsetArr[i]
				if c1 <= ofs.c1 || c2 <= ofs.c2 {
					isTrans = abs(c2-c1) >= abs(ofs.c2-ofs.c1)
					if isTrans {
						trans++
					} else if !ofs.trans {
						ofs.trans = true
						trans++
					}
					break
				}
				if c1 > ofs.c2 && c2 > ofs.c1 {
					offsetArr = append(offsetArr[:i], offsetArr[i+1:]...)
				} else {
					i++
				}
			}
			offsetArr = append(offsetArr, offset{c1: c1, c2: c2, trans: isTrans})
		} else {
			lcss += localCS
			localCS = 0
			if c1 != c2 {
				m := min(c1, c2)
				c1, c2 = m, m
			}
			for j := 0; j < sift4MaxOffset && (c1+j < l1 || c2+j < l2); j++ {
				if c1+j < l1 && s1[c1+j] == s2[c2] {
					c1 += j - 1
					c2--
					break
				}
				if c2+j < l2 && s1[c1] == s2[c2+j] {
					c1--
					c2 += j - 1
					break
				}
			}
		}
		c1++
		c2++
		if c1 >= l1 || c2 >= l2 {
			lcss += localCS
			localCS = 0
			m := min(c1, c2)
			c1, c2 = m, m
		}
	}
	lcss += localCS
	return max(l1, l2) - lcss + trans
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
