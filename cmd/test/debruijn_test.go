package test

import (
	"fmt"
	"testing"
)

func TestMSBEqual(t *testing.T) {
	var d []int = []int{1, 2, 3, 4, 5, 5}
	t.Log(findDuplicate2(d))
}

func findDuplicate(nums []int) int {
	n := len(nums)
	l, r := 1, n-1
	ans := -1
	for l <= r {
		mid := (l + r) >> 1
		fmt.Println("mid", l, mid, r)
		cnt := 0
		for i := 0; i < n; i++ {
			if nums[i] <= mid {
				cnt++
			}
		}

		if cnt <= mid {
			l = mid + 1
		} else {
			r = mid - 1
			ans = mid
		}

		fmt.Println("cnt", cnt, l, r, ans)
	}
	return ans
}

func findDuplicate2(nums []int) int {
	slow, fast := 0, 0
	for slow, fast = nums[slow], nums[nums[fast]]; slow != fast; slow, fast = nums[slow], nums[nums[fast]] {
	}
	slow = 0

	for slow != fast {
		slow = nums[slow]
		fast = nums[fast]
	}

	return slow
}
